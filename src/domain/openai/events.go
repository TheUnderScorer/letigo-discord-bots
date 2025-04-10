package openai

// MemoryUpdated represents an event when a memory has been updated, containing Discord thread ID, vector file ID, and content.
type MemoryUpdated struct {
	DiscordThreadID string
	VectorFileID    string
	Content         string
}
