package chat

import (
	"app/discord"
	chatevents "app/domain/chat/events"
	"app/events"
	"app/llm"
	"app/logging"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

const batchCount = 10

var log = logging.Get().Named("chat").Named("memory")

type DiscordChatMemory struct {
	// messages store the current batch of messages that will be parsed and remembered when length is > batchCount
	messages     []*discordgo.Message
	session      *discordgo.Session
	llmContainer *llm.Container

	inactivityTimer    *time.Timer
	inactivityDuration time.Duration
	mu                 sync.Mutex
}

// TODO also trigger after last message was sent ~30 minutes ago
func NewDiscordChatMemory(session *discordgo.Session, llmContainer *llm.Container) *DiscordChatMemory {
	return &DiscordChatMemory{
		session:            session,
		messages:           []*discordgo.Message{},
		llmContainer:       llmContainer,
		inactivityDuration: 30 * time.Minute,
	}
}

func (m *DiscordChatMemory) StartTick() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.inactivityTimer = time.AfterFunc(m.inactivityDuration, func() {
		log.Info("No messages received for 30 minutes, triggering ForceRemember")

		err := m.ForceRemember()
		if err != nil {
			log.Error("failed to force remember after inactivity period", zap.Error(err))
		}
	})

	log.Info("Started inactivity timer for 30 minutes")
}

func (m *DiscordChatMemory) StopTick() {
	if m.inactivityTimer != nil {
		m.inactivityTimer.Stop()
		log.Info("Stopped inactivity timer")
	}
}

// AddMessage adds a Discord message to the internal memory and triggers memory retention if the batch limit is exceeded.
func (m *DiscordChatMemory) AddMessage(message *discordgo.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)

	// Reset the inactivity timer when a new message arrives
	if m.inactivityTimer != nil {
		m.resetTimer()
		log.Info("Inactivity timer reset due to new message")
	} else {
		// Start tick after the first message is added
		go m.StartTick()
	}

	if len(m.messages) > batchCount {
		err := m.remember()
		if err != nil {
			log.Error("failed to remember messages", zap.Error(err))
		}
	}
}

func (m *DiscordChatMemory) resetTimer() {
	if m.inactivityTimer != nil {
		m.inactivityTimer.Reset(m.inactivityDuration)
	}
}

// ForceRemember triggers a manual memory retention, processing the current batch of messages stored in memory.
func (m *DiscordChatMemory) ForceRemember(additionalMessages ...*discordgo.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, additionalMessages...)

	if len(m.messages) == 0 {
		log.Warn("unable to force remember, no messages.")
		return nil
	}

	return m.remember()
}

func (m *DiscordChatMemory) resetMessages() {
	m.messages = []*discordgo.Message{}
}

// remember processes and commits the current batch of Discord messages into memory
func (m *DiscordChatMemory) remember() error {
	if len(m.messages) == 0 {
		return nil
	}

	// After memory is handled, reset the inactivity timer
	defer m.resetTimer()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get the thread ID from the first message (they should all be from the same thread)
	threadID := m.messages[0].ChannelID

	// Prepare user messages for extraction
	userMessages := make([]string, 0, len(m.messages))
	for _, msg := range m.messages {
		// Skip bot messages
		if msg.Author.Bot {
			continue
		}

		var username string
		friend, ok := discord.Friends[msg.Author.ID]
		if ok {
			username = friend.FirstName
		} else {
			username = msg.Author.Username
		}
		userMessages = append(userMessages, fmt.Sprintf("%s wrote: %s", username, msg.Content))
	}

	if len(userMessages) == 0 {
		log.Info("no user messages to remember")
		return nil
	}

	details, err := m.extractDetails(ctx, strings.Join(userMessages, "\n"), threadID)
	if err != nil {
		log.Error("failed to extract details", zap.Error(err))

		return err
	}
	if details == "" {
		log.Info("no details extracted")
		m.resetMessages()
		return nil
	}

	filteredDetails, err := m.filterDetails(ctx, details, threadID)
	if err != nil {
		log.Error("failed to filter details", zap.Error(err))
		return err
	}
	if filteredDetails == "" {
		log.Info("no details after filtering")
		m.resetMessages()
		return nil
	}

	err = events.Dispatch(ctx, chatevents.MemoryDetailsExtracted{
		Details:         filteredDetails,
		DiscordThreadID: threadID,
	})
	if err != nil {
		log.Error("failed to dispatch MemoryDetailsExtracted event", zap.Error(err), zap.String("threadID", threadID))
		// Don't return here, we still want to clear the messages
	}

	m.resetMessages()

	return nil
}

// filterDetails checks if the extracted details contain information that is already known
// and filters out any duplicate information, returning only new details.
func (m *DiscordChatMemory) filterDetails(ctx context.Context, details string, threadID string) (string, error) {
	// If details are empty, no need to filter
	if details == "" {
		return "", nil
	}

	log.Info("filtering details for duplicates", zap.String("threadID", threadID), zap.String("details", details))

	// Create system prompt for filtering
	systemPrompt := "You are a memory management assistant that identifies new vs. already known information. " +
		"Your task is to analyze details extracted from a conversation and determine which details are genuinely new. " +
		"If you already know certain information from your knowledge base, remove it from the output. " +
		"Return ONLY the details that appear to be new information. " +
		"If all details are already known, return an empty string. " +
		"If all details are new, return them in their original form. " +
		"Format your response as a list of facts separated by newlines." +
		"Please filter out any information you already have in your knowledge base and return only new details" +
		"If there are no new details, return ONLY an empty string"

	// Create a prompt for the LLM
	prompt := llm.Prompt{
		Traits: systemPrompt,
		Phrase: "These details were extracted from a Discord conversation in thread " + threadID + ":\n\n" + details,
	}

	// Call the expensive API to filter details
	response, _, err := m.llmContainer.ExpensiveAPI.Prompt(ctx, prompt)
	if err != nil {
		log.Error("failed to filter details", zap.Error(err), zap.String("threadID", threadID))
		return details, err // Return original details on error
	}

	// If the response is empty or nil, assume no new details
	if response == nil || response.Reply == "" {
		log.Info("no new details after filtering", zap.String("threadID", threadID))
		return "", nil
	}

	// Trim and clean up the response
	filteredDetails := strings.TrimSpace(response.Reply)

	// If the filtered response indicates all details are known, return empty string
	if strings.Contains(strings.ToLower(filteredDetails), "already known") ||
		strings.Contains(strings.ToLower(filteredDetails), "no new details") {
		log.Info("all details already known", zap.String("threadID", threadID))
		return "", nil
	}

	log.Info("filtered details", zap.String("filteredDetails", filteredDetails), zap.String("threadID", threadID))
	return filteredDetails, nil
}

func (m *DiscordChatMemory) extractDetails(ctx context.Context, userMessages string, threadID string) (string, error) {
	systemPrompt := "You are a memory management assistant that parses conversation messages and extracts useful details for memory. " +
		"Extract specific, factual details about people, preferences, or important information mentioned ONLY in these messages. " +
		"Focus on details that answer WHO and WHAT questions. " +
		"Return ONLY the most important details in concise, factual statements separated by newlines. " +
		"If there are no important details worth remembering, return ONLY an empty string." +
		"Do NOT return any details that you already remember." +
		"Do NOT return any additional details, only these extracted from given messages."

	prompt := llm.Prompt{
		Traits: systemPrompt,
		Phrase: "Messages from Discord chat:\n" + userMessages,
	}

	response, _, err := m.llmContainer.ExpensiveAPI.Prompt(ctx, prompt)
	if err != nil {
		log.Error("failed to extract details from messages", zap.Error(err), zap.String("threadID", threadID))
		return "", err
	}

	log.Info("extracted details", zap.String("details", response.Reply), zap.String("threadID", threadID))

	return response.Reply, nil
}
