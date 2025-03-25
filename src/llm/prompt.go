package llm

type Prompt struct {
	// Main prompt send to the LLM
	Phrase string

	// Traits of the LLM, e.g: Be casual
	Traits string
}
