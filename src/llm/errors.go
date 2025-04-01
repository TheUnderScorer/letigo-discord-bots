package llm

import (
	"fmt"
)

type PromptTooLongError struct {
	message string
}

func (p PromptTooLongError) Error() string {
	return p.message
}

func NewPromptTooLongError(length int32) PromptTooLongError {
	return PromptTooLongError{
		message: fmt.Sprintf("Prompt is too long (length %d)", length),
	}
}
