package errors

import (
	"errors"
)

// Error represents an error with some extra information.
type Error interface {
	error

	// WithType adds taints error with "type".
	WithType(t Type) Error
	// WithValue returns an error in which the value associated with `key` is `val`.
	WithValue(key, value interface{}) Error

	getType() Type
	getValue(key interface{}) (interface{}, bool)
}

// Value returns the value for the key, associated with the error.
// Returns nil if no value is associated with key.
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
