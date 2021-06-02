package errors

import (
	"fmt"
)

// Wrap adds context to the error if any. Wrap is meant to be deferred.
// Wrap should be primarily used with named errors.
//
// Example:
//
//	func Foo(a, b string) (rerr error) {
//		defer errors.Wrap(&rerr, "Foo(a=%s, b=%s)", a, b)
//		// do stuff...
//		if err != nil {
//			return err
//		}
//		// continue...
//	}
//
func Wrap(errp *error, format string, a ...interface{}) { //nolint:goprintffuncname
	if *errp == nil {
		return
	}
	s := fmt.Sprintf(format, a...)
	*errp = fmt.Errorf("%s: %w", s, *errp)
}
