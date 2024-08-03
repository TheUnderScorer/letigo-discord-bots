package errors

type UserFriendlyError struct {
	Message string
}

func NewUserFriendlyError(err string) *UserFriendlyError {
	return &UserFriendlyError{Message: err}
}

func (u *UserFriendlyError) Error() string {
	return u.Message
}
