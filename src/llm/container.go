package llm

type Container struct {
	// FreeAPI is a LLM api that is free to use, but in most cases less accurate
	FreeAPI *API
	// ExpensiveAPI is paid, but more accurate
	ExpensiveAPI *API
}
