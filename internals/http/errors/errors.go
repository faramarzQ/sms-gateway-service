package errors

import (
	"errors"
)

type DomainError struct {
	msg string
}

func (e *DomainError) Error() string {
	return e.msg
}

func IsDomainError(err error) bool {
	_, ok := errors.AsType[*DomainError](err)
	return ok
}

var ErrUserNotFound = &DomainError{"user not found"}
var ErrUserBalanceExceeded = &DomainError{"user balance exceeded"}
