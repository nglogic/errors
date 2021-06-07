package errors

import (
	"errors"
	"fmt"
	"reflect"
)

// simpleError represents error with some additional context data.
type simpleError struct {
	e      error
	t      Type
	values map[interface{}]interface{}
}

// New creates new error.
func New(message string) Error {
	return &simpleError{
		e: errors.New(message),
	}
}

// Newf creates new error with message according to format and arguments.
func Newf(format string, args ...interface{}) Error {
	return &simpleError{
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
	return &simpleError{
		e: err,
	}
}

// Error returns error message.
func (e *simpleError) Error() string {
	if e == nil || e.e == nil {
		return ""
	}

	return e.e.Error()
}

// WithType adds taints error with "type".
func (e *simpleError) WithType(t Type) Error {
	if e == nil {
		return nil
	}
	e.t = t
	return e
}

func (e *simpleError) WithValue(key, value interface{}) Error {
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
func (e *simpleError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.e
}

func (e *simpleError) getType() Type {
	return e.t
}

func (e *simpleError) getValue(key interface{}) (interface{}, bool) {
	if v, ok := e.values[key]; ok {
		return v, true
	}
	return nil, false
}
