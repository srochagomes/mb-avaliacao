package app

import "errors"

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrOrderNotFound   = errors.New("order not found")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
