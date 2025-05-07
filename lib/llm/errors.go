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

type ErrRefusedToReply struct {
	reason string
	prompt string
}

func NewRefusedToReplyError(reason string, prompt string) ErrRefusedToReply {
	return ErrRefusedToReply{
		reason: reason,
		prompt: prompt,
	}
}

func (r ErrRefusedToReply) Prompt() string {
	return r.prompt
}

func (r ErrRefusedToReply) Error() string {
	return "Refused to reply: " + r.reason
}
