package chatevents

// MemoryDetailsExtracted represents details extracted from a chat message thread for memory and storage purposes.
// Details contains the extracted relevant information.
// DiscordThreadID is the ID of the Discord thread associated with the memory.
type MemoryDetailsExtracted struct {
	Details         string
	DiscordThreadID string
}
