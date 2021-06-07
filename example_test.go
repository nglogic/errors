package errors_test

import (
	"fmt"
	"log"
	"os"

	"github.com/nglogic/errors"
)

// This example demonstrates the usage of From function.
func ExampleFrom() {
	err := errors.From(os.ErrNotExist).WithType(errors.TypeNotFound)

	fmt.Println(err)
	// Output: file does not exist
}

// This example demonstrates the usage of passing extra data in errors.
// In this case we want to log error, but we want to use other message to display to the user.
func ExampleError_withValue() {
	type errPublicMessageKey struct{}

	f := func() error {
		return errors.New("io failure").
			WithValue(errPublicMessageKey{}, "There was a problem with filesystem, please try again later.")
	}

	err := f()
	if err != nil {
		// Log error internally.
		log.Printf("f failed: %v", err)

		// Hide real problem from user, display them a nice message.
		msg := errors.Value(err, errPublicMessageKey{})
		if msg != nil {
			fmt.Println(msg.(string))
		} else {
			fmt.Println("Internal server error")
		}
	}
	// Output: There was a problem with filesystem, please try again later.
}

// This example demonstrates the usage of Append function.
func ExampleAppend() {
	var err error

	if check1 := false; !check1 {
		err = errors.Append(err, errors.New("check 1 failed"))
	}
	if check2 := false; !check2 {
		err = errors.Append(err, errors.New("check 2 failed"))
	}
	if check3 := false; !check3 {
		err = errors.Append(err, errors.New("check 3 failed"))
	}
	if err != nil {
		err = errors.From(err).WithType(errors.TypeInvalidRequest)
	}

	fmt.Printf(
		"error: '%v', is invalid request: %t",
		err,
		errors.IsType(err, errors.TypeInvalidRequest),
	)
	// Output: error: 'check 1 failed; check 2 failed; check 3 failed', is invalid request: true
}

// This example demonstrates how to simplify error handling in
// a function using Finalize.
func ExampleFinalize() {
	foo := func(a, b int) (rerr error) {
		defer errors.Finalize(&rerr, "task failed (a=%d, b=%d)", a, b)

		if err := doWork(a + b); err != nil {
			return err
		}
		return nil
	}

	err := foo(1, 2)
	fmt.Printf("%v\n", err)

	err = foo(-1, 1)
	fmt.Printf("%v\n", err)

	// Output:
	// <nil>
	// task failed (a=-1, b=1): couldn't do the work with 0 value
}

func doWork(v int) error {
	if v == 0 {
		return errors.New("couldn't do the work with 0 value")
	}
	return nil
}
