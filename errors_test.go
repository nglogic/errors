package errors_test

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/nglogic/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	tests := []struct {
		name        string
		makeErr     func() error
		wantErr     bool
		wantMessage string
		wantType    errors.Type
	}{
		{
			name: "FromErr with nil",
			makeErr: func() error {
				return errors.FromErr(nil)
			},
			wantErr: false,
		},
		{
			name: "simple",
			makeErr: func() error {
				return errors.New("test")
			},
			wantErr:     true,
			wantMessage: "test",
			wantType:    errors.TypeUnknown,
		},
		{
			name: "simple with formatting",
			makeErr: func() error {
				return errors.Newf("test %d %d %d", 1, 2, 3)
			},
			wantErr:     true,
			wantMessage: "test 1 2 3",
			wantType:    errors.TypeUnknown,
		},
		{
			name: "simple with type",
			makeErr: func() error {
				return errors.New("test").WithType(errors.TypeNotFound)
			},
			wantErr:     true,
			wantMessage: "test",
			wantType:    errors.TypeNotFound,
		},
		{
			name: "wapped with type",
			makeErr: func() error {
				err := fmt.Errorf("stderror")
				return errors.FromErr(err).WithType(errors.TypeNotFound)
			},
			wantErr:     true,
			wantMessage: "stderror",
			wantType:    errors.TypeNotFound,
		},
		{
			name: "multiple levels of wrapping with type and labels, mixed std errors with custom errors",
			makeErr: func() error {
				err := fmt.Errorf("stderror")
				err = errors.FromErr(fmt.Errorf("wrap1: %w", err))
				err = fmt.Errorf("wrap%d: %w", 2, err)
				err = errors.FromErr(err)
				err = errors.FromErr(err).WithType(errors.TypeNotFound)
				err = errors.FromErr(err).WithType(errors.TypeInvalidRequest) // This type should take precedence!
				err = fmt.Errorf("wrap3: %w", err)
				return err
			},
			wantErr:     true,
			wantMessage: "wrap3: wrap2: wrap1: stderror",
			wantType:    errors.TypeInvalidRequest,
		},
		{
			name: "append with nils",
			makeErr: func() error {
				return errors.Append(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "append with nils 2",
			makeErr: func() error {
				return errors.Append(errors.Append(nil, nil, nil), nil, errors.Append(nil, nil))
			},
			wantErr: false,
		},
		{
			name:        "append returns not nil if at least one err was added",
			makeErr:     func() error { return errors.Append(nil, errors.New("test")) },
			wantErr:     true,
			wantMessage: "test",
			wantType:    errors.TypeUnknown,
		},
		{
			name: "only non nil errors are appended",
			makeErr: func() error {
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
			name: "err type is preserved by append",
			makeErr: func() error {
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
			name: "err type is preserved by append - multilevel",
			makeErr: func() error {
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
			name: "type set on the result of append has precedence",
			makeErr: func() error {
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
			name: "all features at once with multilevel append",
			makeErr: func() error {
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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.makeErr()
			if !tt.wantErr {
				// Look out for using assert.Nil here! In this case it is not the same. assert.Nil only checks if value of an interface is nil.
				// See: https://yourbasic.org/golang/gotcha-why-nil-error-not-equal-nil/
				// It is crutial to check if simple equality works, because that's how we check errors in the code.
				assert.True(t, err == nil, "error is not nil")
				return
			}
			require.NotNil(t, err)

			assert.Equal(t, tt.wantMessage, err.Error(), "invalid error message")
			assert.Equal(t, tt.wantType, errors.GetType(err), "invalid error type")
			assert.True(t, errors.IsType(err, tt.wantType), "invalid error type chek result")
		})
	}
}

func TestErrorValue(t *testing.T) {
	type valueKey struct{}
	value := "test value"

	tests := []struct {
		name      string
		makeErr   func() error
		wantValue bool
	}{
		{
			name: "standard error",
			makeErr: func() error {
				return fmt.Errorf("test")
			},
			wantValue: false,
		},
		{
			name: "simple with no value",
			makeErr: func() error {
				return errors.New("test")
			},
			wantValue: false,
		},
		{
			name: "simple with value",
			makeErr: func() error {
				return errors.New("test").WithValue(valueKey{}, value)
			},
			wantValue: true,
		},
		{
			name: "nested with value",
			makeErr: func() error {
				return errors.FromErr(
					fmt.Errorf(
						"test: %w",
						errors.New("test").WithValue(valueKey{}, value),
					),
				)
			},
			wantValue: true,
		},
		{
			name: "nested with append, with value",
			makeErr: func() error {
				return errors.FromErr(
					fmt.Errorf(
						"test: %w",
						errors.Append(
							fmt.Errorf("e1"),
							fmt.Errorf(
								"e2: %w",
								errors.Append(
									fmt.Errorf("ee1"),
									fmt.Errorf(
										"ee2: %w",
										errors.New("ee3").WithValue(valueKey{}, value),
									),
								),
							),
						),
					),
				)
			},
			wantValue: true,
		},
		{
			name: "nested with append, with value 2",
			makeErr: func() error {
				return errors.FromErr(
					fmt.Errorf(
						"test: %w",
						errors.Append(
							fmt.Errorf("e1"),
							fmt.Errorf(
								"e2: %w",
								errors.Append(
									fmt.Errorf("ee1"),
									fmt.Errorf(
										"ee2: %w",
										errors.New("ee3"),
									),
								),
							),
						).WithValue(valueKey{}, value),
					),
				)
			},
			wantValue: true,
		},
		{
			name: "nested with append, with value 3",
			makeErr: func() error {
				return errors.FromErr(
					fmt.Errorf(
						"test: %w",
						errors.Append(
							fmt.Errorf("e1"),
							errors.Append(
								fmt.Errorf("ee1"),
								fmt.Errorf(
									"ee2: %w",
									errors.New("eee1"),
								),
							),
							fmt.Errorf(
								"e3: %w",
								errors.Append(
									fmt.Errorf("ee1"),
									fmt.Errorf(
										"ee2: %w",
										errors.New("eee1"),
									),
								).WithValue(valueKey{}, value),
							),
						),
					),
				)
			},
			wantValue: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.makeErr()
			require.NotNil(t, err)

			v := errors.Value(err, valueKey{})
			if !tt.wantValue {
				assert.True(t, v == nil)
				return
			}
			require.NotNil(t, v)
			assert.Equal(t, v, value)
		})
	}
}

func TestWithFinalContext(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		err := func() (rerr error) {
			defer errors.WithFinalContext(&rerr, "failed to stf")
			return nil
		}()
		assert.True(t, err == nil)
	})
	t.Run("error is wrapped with context", func(t *testing.T) {
		err := func() (rerr error) {
			defer errors.WithFinalContext(&rerr, "foo failed with a=%d", 123)
			return io.ErrClosedPipe
		}()
		assert.Error(t, err)
		assert.Equal(
			t,
			fmt.Sprintf("foo failed with a=123: %s", io.ErrClosedPipe.Error()),
			err.Error(),
		)
	})
	t.Run("wrapped error preserves type", func(t *testing.T) {
		err := func() (rerr error) {
			defer errors.WithFinalContext(&rerr, "func failed")
			return errors.New("not found").WithType(errors.TypeNotFound)
		}()
		assert.Error(t, err)
		assert.True(t, errors.IsType(err, errors.TypeNotFound))
	})
}

func TestIs(t *testing.T) {
	tests := []struct {
		name    string
		makeErr func() error
		want    error
	}{
		{
			name: "fromErr",
			makeErr: func() error {
				return errors.FromErr(os.ErrNotExist)
			},
			want: os.ErrNotExist,
		},
		{
			name: "fromErr with context",
			makeErr: func() error {
				return fmt.Errorf("test: %w", errors.FromErr(os.ErrNotExist))
			},
			want: os.ErrNotExist,
		},
		{
			name: "fromErr with context 2",
			makeErr: func() error {
				return errors.FromErr(fmt.Errorf("test: %w", errors.FromErr(os.ErrNotExist)))
			},
			want: os.ErrNotExist,
		},
		{
			name: "append",
			makeErr: func() error {
				err := errors.Append(nil, errors.New("test"))
				err = errors.Append(err, os.ErrNotExist)
				return err
			},
			want: os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.makeErr()
			assert.True(t, errors.Is(err, tt.want))
		})
	}
}

type testError struct{ Value string }

func (s testError) Error() string { return s.Value }

func TestAs(t *testing.T) {
	tests := []struct {
		name    string
		makeErr func() error
	}{
		{
			name: "fromErr",
			makeErr: func() error {
				return errors.FromErr(&testError{Value: "test"})
			},
		},
		{
			name: "fromErr with context",
			makeErr: func() error {
				return fmt.Errorf("test: %w", errors.FromErr(&testError{Value: "test"}))
			},
		},
		{
			name: "fromErr with context 2",
			makeErr: func() error {
				return errors.FromErr(fmt.Errorf("test: %w", errors.FromErr(&testError{Value: "test"})))
			},
		},
		{
			name: "append",
			makeErr: func() error {
				err := errors.Append(nil, errors.New("test"))
				err = errors.Append(err, &testError{Value: "test"})
				return err
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.makeErr()

			var ei *testError
			ok := errors.As(err, &ei)
			assert.True(t, ok)
		})
	}
}
