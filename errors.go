// Package errors defines the error type, which carries type information
// and extra values for usage across API boundaries.
// The package provides Append function for representing a list of errors as a single error.
package errors

import (
	"errors"
	"fmt"
	"reflect"
)

// Error represents an error with some extra information.
type Error interface {
	error

	// WithType adds taints error with "type".
	WithType(t Type) Error
	// WithValue returns an error in which the value associated with `key` is `val`.
	// A key can be any type that supports equality.
	WithValue(key string, value interface{}) Error

	getType() Type
	getValue(key string) (interface{}, bool)
}

// Value returns the value associated with this error for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
//
// A key can be any type that supports equality.
// Packages should define keys as an unexported type to avoid collisions.
func Value(err error, key string) interface{} {
	for err != nil {
		if intErr, ok := err.(Error); ok {
			if v, ok := intErr.getValue(key); ok {
				return v
			}
		}
		err = errors.Unwrap(err)
	}

	return nil
}

// Finalize is a helper function that works nicely with defer and named returns.
// It adds context to the error if it's not nil.
func Finalize(ep *error, format string, a ...interface{}) {
	if *ep == nil {
		return
	}
	*ep = fmt.Errorf("%s: %w", fmt.Sprintf(format, a...), *ep)
}

// As is a convenient alias for standard errors package function, that helps to avoid extra imports.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is is a convenient alias for standard errors package function, that helps to avoid extra imports.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Unwrap is a convenient alias for standard errors package function, that helps to avoid extra imports.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// internalError represents error with some additional context data.
type internalError struct {
	e      error
	t      Type
	values map[interface{}]interface{}
}

// New creates new error.
func New(message string) Error {
	return &internalError{
		e: errors.New(message),
	}
}

// Newf creates new error with message according to format and arguments.
func Newf(format string, args ...interface{}) Error {
	return &internalError{
		e: fmt.Errorf(format, args...),
	}
}

// From takes existing error and creates internal error.
// This allows taking standard error and adding some details to it.
//
// If passed error is nil, returns nil.
func From(err error) Error {
	if err == nil {
		return nil
	}
	return &internalError{
		e: err,
	}
}

// Error returns error message.
func (e *internalError) Error() string {
	if e == nil || e.e == nil {
		return ""
	}

	return e.e.Error()
}

// WithType adds taints error with "type".
func (e *internalError) WithType(t Type) Error {
	if e == nil {
		return nil
	}
	e.t = t
	return e
}

func (e *internalError) WithValue(key string, value interface{}) Error {
	if e == nil {
		return nil
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	if e.values == nil {
		e.values = make(map[interface{}]interface{})
	}
	e.values[key] = value
	return e
}

// Unwrap implements std errors package unwrapping.
// See https://golang.org/pkg/errors/#Unwrap.
func (e *internalError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.e
}

func (e *internalError) getType() Type {
	return e.t
}

func (e *internalError) getValue(key string) (interface{}, bool) {
	if v, ok := e.values[key]; ok {
		return v, true
	}
	return nil, false
}
