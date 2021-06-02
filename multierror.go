package errors

import (
	"errors"
	"reflect"
	"strings"
)

// multiError allows to store multiple errors as one.
//
// Errors aggregated in this object will retain properties supported in this package: labels, public labels, type.
// It is suitable for cases when process may result in multiple errors. Alternatively, if you don't want to fail on the
// first error encounter and error creation is independent. This way you could return information about all problems.
type multiError struct {
	errs   []error
	t      Type
	values map[interface{}]interface{}
}

// Append appends errors to the existing error to create one, aggregated error.
//
// If all the passed errors are nil, returns nil.
func Append(to error, errors ...error) Error {
	switch to := to.(type) {
	case *multiError:
		var es []error
		for _, e := range errors {
			if e == nil {
				continue
			}
			es = append(es, e)
		}
		if len(es) == 0 {
			return to
		}

		if to == nil {
			to = &multiError{}
		}
		to.errs = append(to.errs, es...)

		return to
	default:
		var es []error
		if to != nil {
			es = append(es, to)
		}
		for _, e := range errors {
			if e == nil {
				continue
			}
			es = append(es, e)
		}
		if len(es) == 0 {
			return nil
		}

		return &multiError{errs: es}
	}
}

func (e *multiError) WithType(t Type) Error {
	if e == nil {
		return nil
	}

	e.t = t
	return e
}

func (e *multiError) Error() string {
	if e == nil {
		return ""
	}

	messages := make([]string, 0, len(e.errs))
	for _, err := range e.errs {
		message := err.Error()
		if message != "" {
			messages = append(messages, message)
		}
	}
	return strings.Join(uniqueStrings(messages), "; ")
}

func (e *multiError) WithValue(key, value interface{}) Error {
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

func (e *multiError) getType() Type {
	if e == nil {
		return TypeUnknown
	}
	if e.t != TypeUnknown {
		return e.t
	}

	var et Type
	for _, err := range e.errs {
		if t := GetType(err); t != TypeUnknown {
			et = t
		}
	}
	return et
}

func (e *multiError) getValue(key interface{}) (interface{}, bool) {
	if v, ok := e.values[key]; ok {
		return v, true
	}

	for _, se := range e.errs {
		for se != nil {
			if intErr, ok := se.(Error); ok {
				if v, ok := intErr.getValue(key); ok {
					return v, true
				}
			}
			se = errors.Unwrap(se)
		}
	}
	return nil, false
}

func (e multiError) As(target interface{}) bool {
	for _, err := range e.errs {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

func (e multiError) Is(target error) bool {
	for _, err := range e.errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func uniqueStrings(ss []string) []string {
	tmp := make(map[string]bool)
	for _, s := range ss {
		tmp[s] = true
	}

	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if !tmp[s] {
			continue
		}
		tmp[s] = false
		out = append(out, s)
	}
	return out
}
