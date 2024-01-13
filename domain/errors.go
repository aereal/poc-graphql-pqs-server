package domain

import "fmt"

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
