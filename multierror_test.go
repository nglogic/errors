package errors_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/nglogic/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testError struct{ Value string }

func (s testError) XD() {}

func (s testError) Error() string { return "" }

// nolint:funlen // This test is extensive :)
func TestMulti(t *testing.T) {
	cases := []struct {
		name        string
		makeErr     func(t *testing.T) error
		wantErr     bool
		wantMessage string
		wantType    errors.Type
	}{
		{
			name: "don't panic if added err is nil",
			makeErr: func(t *testing.T) error {
				return errors.Append(nil, nil)
			},
			wantErr: false,
		},
		{
			name:        "not nil if at least one err was added",
			makeErr:     func(t *testing.T) error { return errors.Append(nil, errors.New("xD")) },
			wantErr:     true,
			wantMessage: "xD",
			wantType:    errors.TypeUnknown,
		},
		{
			name: "only non nil errors are appended",
			makeErr: func(t *testing.T) error {
				return errors.Append(
					errors.New("test"),
					nil,
					errors.New("test2"),
					nil,
					errors.New("test3"),
				)
			},
			wantErr:     true,
			wantMessage: "test; test2; test3",
			wantType:    errors.TypeUnknown,
		},
		{
			name: "err type is preserved",
			makeErr: func(t *testing.T) error {
				return errors.Append(
					errors.New("").WithType(errors.TypeInvalidRequest),
					fmt.Errorf(""),
					errors.New("").WithType(errors.TypeAlreadyExists), // The last one that should be chosen.
				)
			},
			wantErr:     true,
			wantMessage: "",
			wantType:    errors.TypeAlreadyExists,
		},
		{
			name: "err type is preserved - nested types",
			makeErr: func(t *testing.T) error {
				return errors.Append(
					errors.New("").WithType(errors.TypeInvalidRequest),
					fmt.Errorf(""),
					errors.New("").WithType(errors.TypeAlreadyExists),
					errors.Append( // The last one that should be chosen. But...
						errors.New(""),
						errors.New("").WithType(errors.TypePermissionDenied), // ...parent inherits the type from here.
						errors.New(""),
					),
				)
			},
			wantErr:     true,
			wantMessage: "",
			wantType:    errors.TypePermissionDenied,
		},
		{
			name: "type set on multierror has precedence",
			makeErr: func(t *testing.T) error {
				return errors.Append(
					errors.New("").WithType(errors.TypeInvalidRequest),
					fmt.Errorf(""),
					errors.New("").WithType(errors.TypeAlreadyExists), // The last one that should be chosen.
				).WithType(errors.TypeNotFound)
			},
			wantErr:     true,
			wantMessage: "",
			wantType:    errors.TypeNotFound,
		},
		{
			name: "all features at once with nested `Append`",
			makeErr: func(t *testing.T) error {
				return errors.Append(
					errors.New("invalid attribute 1"),
					errors.New("invalid attribute 2").WithType(errors.TypeFailedPrecondition),
					errors.Append(
						errors.New("attribute 3 is too big"),
						errors.New("attribute 3 has invalid format"),
					),
					errors.New("missing attribute 4"),
				).WithType(errors.TypeInvalidRequest)
			},
			wantErr:     true,
			wantMessage: "invalid attribute 1; invalid attribute 2; attribute 3 is too big; attribute 3 has invalid format; missing attribute 4",
			wantType:    errors.TypeInvalidRequest,
		},

		// Examples
		{
			name: "example usage: prepare error to return in case of any of other errors is not nil, but everything went ok",
			makeErr: func(t *testing.T) error {
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

				return err
			},
			wantErr: false,
		},
		{
			name: "example usage: prepare error to return in case of any of other errors is not nil, and there was a failure",
			makeErr: func(t *testing.T) error {
				var err error

				if check1 := true; !check1 {
					err = errors.Append(err, errors.New("check 1 failed").WithType(errors.TypeInvalidRequest))
				}
				if check2 := false; !check2 { // <- This one will fail!
					err = errors.Append(err, errors.New("check 2 failed").WithType(errors.TypeInvalidRequest))
				}
				if check3 := true; !check3 {
					err = errors.Append(err, errors.New("check 3 failed").WithType(errors.TypeFailedPrecondition))
				}
				if check4 := false; !check4 { // <- This one will fail!
					err = errors.Append(err, errors.New("check 4 failed").WithType(errors.TypeInvalidRequest))
				}

				return err
			},
			wantErr:     true,
			wantMessage: "check 2 failed; check 4 failed",
			wantType:    errors.TypeInvalidRequest,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.makeErr(t)
			if !tc.wantErr {
				// Look out for using assert.Nil here! In this case it is not the same. assert.Nil only checks if value of an interface is nil.
				// See: https://yourbasic.org/golang/gotcha-why-nil-error-not-equal-nil/
				// It is crutial to check if simple equality works, because that's how we check errors in the code.
				assert.True(t, err == nil, "error is not nil")
				return
			}
			require.NotNil(t, err)

			assert.Equal(t, tc.wantMessage, err.Error())
			assert.Equal(t, tc.wantType, errors.GetType(err))

			// Check if global functions for types and groups work with multierror properly.
			tp := errors.GetType(err)
			assert.True(t, errors.IsType(err, tp))
			assert.True(t, errors.IsGroup(err, tp.Group))
			assert.Equal(t, tp.Group, errors.GetGroup(err))
		})
	}
}

func TestMultiIs(t *testing.T) {
	err := errors.Append(nil, errors.New("something went wrong"))
	err = errors.Append(err, os.ErrNotExist)
	isNotExist := errors.Is(err, os.ErrNotExist)
	assert.True(t, isNotExist)
}

func TestMultiAs(t *testing.T) {
	var ei *testError

	err := errors.Append(nil, os.ErrNotExist)
	err = errors.Append(err, &testError{Value: "xD"})

	errFound := errors.As(err, &ei)
	assert.True(t, errFound)
	assert.Equal(t, ei.Value, "xD")
}

func TestNilMultiErr(t *testing.T) {
	err := errors.Append(nil, nil, nil)
	assert.True(t, err == nil)

	err = errors.Append(errors.Append(nil, nil, nil), nil, errors.Append(nil, nil))
	assert.True(t, err == nil)
}
