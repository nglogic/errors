package errors

import (
	"errors"
)

// Predefined groups.
const (
	GroupServer Group = iota
	GroupClient
)

// Predefined types.
//
// If you need to use a custom type you can define it yourself.
//
//   const MyTypeResourceExhausted = Type{Name: "ResourceExhausted ", Group: GroupServer}
//
var (
	TypeUnknown = Type{}

	TypeDeadlineExceeded = Type{Name: "DeadlineExceeded", Group: GroupServer}

	TypeInvalidRequest     = Type{Name: "InvalidRequest", Group: GroupClient}
	TypeAlreadyExists      = Type{Name: "AlreadyExists", Group: GroupClient}
	TypeNotFound           = Type{Name: "NotFound", Group: GroupClient}
	TypeUnauthenticated    = Type{Name: "Unauthenticated", Group: GroupClient}
	TypePermissionDenied   = Type{Name: "PermissionDenied", Group: GroupClient}
	TypeFailedPrecondition = Type{Name: "FailedPrecondition", Group: GroupClient}
)

// Group represent errors groups.
// The primary purpose is to distinguish between client and server errors, but you can create custom groups, if needed.
type Group int

func (g Group) String() string {
	switch g {
	case GroupClient:
		return "Client"
	default:
		return "Server"
	}
}

// Type describes general error type.
// Packs information about status codes and group into uint64.
type Type struct {
	Name  string
	Group Group
}

func (t Type) String() string {
	if t.Name == "" {
		return "Unknown"
	}
	return t.Name
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

// IsGroup checks if error belongs to given "group".
func IsGroup(err error, g Group) bool {
	for err != nil {
		if e, ok := err.(Error); ok {
			if t := e.getType(); t != TypeUnknown {
				if t.Group == g {
					return true
				}
			}
		}

		err = errors.Unwrap(err)
	}

	// If no information is available, assume this is servier's fault.
	return g == GroupServer
}

// GetGroup returns error's group.
// If there's no group information in error, assumes this is servier's fault and returns GroupServer.
// If there are multiple groups in error chain, returns most recent one.
func GetGroup(err error) Group {
	for err != nil {
		if e, ok := err.(Error); ok {
			if t := e.getType(); t != TypeUnknown {
				return t.Group
			}
		}

		err = errors.Unwrap(err)
	}

	// If no information is available, assume this is server's fault.
	return GroupServer
}
