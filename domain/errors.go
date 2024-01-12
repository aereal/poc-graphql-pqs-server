package domain

import "fmt"

type QueryBuildError struct{ err error }

func (e *QueryBuildError) Error() string { return fmt.Sprintf("failed to build query: %s", e.err) }

func (e *QueryBuildError) Unwrap() error { return e.err }
