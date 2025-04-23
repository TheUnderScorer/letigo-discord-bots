package llm

import (
	"fmt"
)

type ErrPromptTooLong struct {
	message string
}

func (p ErrPromptTooLong) Error() string {
	return p.message
}

func NewPromptTooLongError(length int32) ErrPromptTooLong {
	return ErrPromptTooLong{
		message: fmt.Sprintf("Prompt is too long (length %d)", length),
	}
}
