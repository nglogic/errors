package errors_test

import (
	"fmt"

	"github.com/nglogic/errors"
)

func ExampleAppend_nilErr() {
	var err error

	if check1 := true; !check1 {
		err = errors.Append(err, errors.New("check 1 failed"))
	}
	if check2 := true; !check2 {
		err = errors.Append(err, errors.New("check 2 failed"))
	}
	if check3 := true; !check3 {
		err = errors.Append(err, errors.New("check 3 failed"))
	}
	if err != nil {
		err = errors.FromErr(err).WithType(errors.TypeInvalidRequest)
	}

	fmt.Printf("%v", err)
	// Output: <nil>
}

func ExampleAppend_nonNilErr() {
	var err error

	if check1 := true; !check1 {
		err = errors.Append(err, errors.New("check 1 failed"))
	}
	if check2 := false; !check2 {
		err = errors.Append(err, errors.New("check 2 failed"))
	}
	if check3 := false; !check3 {
		err = errors.Append(err, errors.New("check 3 failed"))
	}
	if err != nil {
		err = errors.FromErr(err).WithType(errors.TypeInvalidRequest)
	}

	fmt.Printf(
		"error: '%v', is invalid request: %t",
		err,
		errors.IsType(err, errors.TypeInvalidRequest),
	)
	// Output: error: 'check 2 failed; check 3 failed', is invalid request: true
}
