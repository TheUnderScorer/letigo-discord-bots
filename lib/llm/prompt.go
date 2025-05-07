package llm

type Prompt struct {
	// Main prompt send to the LLM
	Phrase string

	// Traits of the LLM, e.g: Be casual
	Traits string

	Files []File
}

func NewPrompt(phrase string) *Prompt {
	return &Prompt{
		Phrase: phrase,
	}
}
