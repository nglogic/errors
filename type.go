package errors

import (
	"errors"
)

// Predefined types.
//
// If you need to use a custom type you can define it yourself.
//
//   const MyTypeResourceExhausted errors.Type = "ResourceExhausted"
//
const (
	TypeUnknown            Type = ""
	TypeDeadlineExceeded   Type = "DeadlineExceeded"
	TypeInvalidRequest     Type = "InvalidRequest"
	TypeAlreadyExists      Type = "AlreadyExists"
	TypeNotFound           Type = "NotFound"
	TypeUnauthenticated    Type = "Unauthenticated"
	TypePermissionDenied   Type = "PermissionDenied"
	TypeFailedPrecondition Type = "FailedPrecondition"
)

// Type represents general error type.
type Type string

func (t Type) String() string {
	if t == "" {
		return "Unknown"
	}
	return string(t)
}

// IsType checks if error is tainted with given "type".
func IsType(err error, t Type) bool {
	for err != nil {
		if e, ok := err.(Error); ok {
			if et := e.getType(); t != TypeUnknown {
				if et == t {
					return true
				}
			}
		}

		err = errors.Unwrap(err)
	}

	return t == TypeUnknown
}

// GetType returns error's type.
// If there's no type information in error, returns TypeUnknown.
// If there are multiple types in error chain, returns most recent one.
func GetType(err error) Type {
	for err != nil {
		if e, ok := err.(Error); ok {
			if t := e.getType(); t != TypeUnknown {
				return t
			}
		}

		err = errors.Unwrap(err)
	}

	return TypeUnknown
}
