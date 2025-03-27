package errors

// PublicError defines error struct that can be exposed to users
type PublicError struct {
	// Message is an error message, it should be user-friendly
	Message string `json:"message,omitempty"`
	// Cause is an underlying cause of the error
	Cause error `json:"cause,omitempty"`
	// Context is an optional error helpful in debugging the error
	Context map[string]any `json:"context,omitempty"`
}

func NewPublicError(err string) *PublicError {
	return &PublicError{Message: err, Context: make(map[string]any)}
}

func NewPublicErrorCause(err string, cause error) *PublicError {
	return &PublicError{Message: err, Cause: cause, Context: make(map[string]any)}
}

func (u *PublicError) Error() string {
	return u.Message
}

func (u *PublicError) AddContext(key string, value any) {
	u.Context[key] = value
}
