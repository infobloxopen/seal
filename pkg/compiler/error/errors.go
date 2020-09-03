package compiler_error

import (
	"errors"
	"fmt"
)

// Errors ...
var (
	ErrEmptyLanguage    = errors.New("invalid empty language")
	ErrEmptyPolicies    = errors.New("invalid empty policies")
	ErrEmptySubject     = errors.New("invalid empty subject")
	ErrInvalidSubject   = errors.New("invalid invalid subject")
	ErrEmptyVerb        = errors.New("invalid empty verb")
	ErrEmptyTypePattern = errors.New("invalid empty type-pattern")
	ErrUnknownSelector  = errors.New("invalid unknown selector")
)

// Error defines a compiler specific error type
type Error struct {
	Err  error
	Line int
	Desc string
}

// Error satisfies the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("compiler_rego: at #%d %s due to error: %s", e.Line, e.Desc, e.Err)
}

// New is a convenience function to create an Error
func New(err error, line int, desc string) *Error {
	return &Error{
		Err:  err,
		Line: line,
		Desc: desc,
	}
}
