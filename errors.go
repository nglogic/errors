// Package errors defines the error type, which carries type information
// and extra values for usage across API boundaries.
// The package provides Append function for representing a list of errors as a single error.
package errors

import (
	"errors"
	"fmt"
)

// Error represents an error with some extra information.
type Error interface {
	error

	// WithType adds taints error with "type".
	WithType(t Type) Error
	// WithValue returns an error in which the value associated with `key` is `val`.
	// A key can be any type that supports equality.
	WithValue(key, value interface{}) Error

	getType() Type
	getValue(key interface{}) (interface{}, bool)
}

// Value returns the value associated with this error for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
//
// A key can be any type that supports equality.
// Packages should define keys as an unexported type to avoid collisions.
func Value(err error, key interface{}) interface{} {
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

// WithFinalContext is a helper function that works nicely with defer.
// It adds context to the error if it's not nil.
// It is designed to work with named errors.
func WithFinalContext(ep *error, format string, a ...interface{}) {
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
