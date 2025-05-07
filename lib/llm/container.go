package llm

type Container struct {
	// FreeAPI is a LLM api that is free to use, but in most cases less accurate
	FreeAPI *API
	// AssistantAPI is paid API, with custom instructions.
	AssistantAPI *API
	// ExpensiveAPI is a variant of AssistantAPI, but without custom instructions
	ExpensiveAPI *API
}
