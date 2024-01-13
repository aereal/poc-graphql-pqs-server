package domain

import (
	"errors"
	"fmt"
)

type QueryBuildError struct{ err error }

func (e *QueryBuildError) Error() string { return fmt.Sprintf("failed to build query: %s", e.err) }

func (e *QueryBuildError) Unwrap() error { return e.err }

type NotFoundError[K comparable, T any] struct {
	Key K
}

func (e *NotFoundError[K, T]) Error() string {
	var t T
	return fmt.Sprintf("%T (key: %v (%T)) is not found", t, e.Key, e.Key)
}

var (
	ErrInvalidOrderDirection      = errors.New("invalid order direction")
	ErrInvalidCharacterOrderField = errors.New("invalid charcter order field")
	ErrInvalidLimit               = errors.New("invalid limit")
	ErrUnknownNumericKind         = errors.New("unknown numeric kind")
)
