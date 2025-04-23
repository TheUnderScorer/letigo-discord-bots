package errors

import "errors"

// ErrPublic defines error struct that can be exposed to users
type ErrPublic struct {
	// Message is an error message, it should be user-friendly
	Message string `json:"message,omitempty"`
	// Cause is an underlying cause of the error
	Cause error `json:"cause,omitempty"`
	// Context is an optional error helpful in debugging the error
	Context map[string]any `json:"context,omitempty"`
}

func NewErrPublic(err string) *ErrPublic {
	return &ErrPublic{Message: err, Context: make(map[string]any)}
}

func NewErrPublicCause(err string, cause error) error {
	return errors.Join(cause, &ErrPublic{Message: err, Cause: cause, Context: make(map[string]any)})
}

func (u *ErrPublic) Error() string {
	return u.Message
}

func (u *ErrPublic) AddContext(key string, value any) {
	u.Context[key] = value
}
