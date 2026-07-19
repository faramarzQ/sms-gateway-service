package errors

import "errors"

type DomainError error

var ErrUserNotFound DomainError = errors.New("user not found")
var ErrUserBalanceExceeded DomainError = errors.New("user balance exceeded")
