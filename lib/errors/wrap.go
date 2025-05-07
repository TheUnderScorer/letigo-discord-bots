package errors

import e "errors"

// Wrap returns a new error that wraps the given error with an additional message for added context.
func Wrap(err error, msg string) error {
	return e.Join(e.New(msg), err)
}
