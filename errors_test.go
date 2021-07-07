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
			name: "From with nil",
			makeErr: func() error {
				return errors.From(nil)
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
			name: "wrapped with type",
			makeErr: func() error {
				err := fmt.Errorf("stderror")
				return errors.From(err).WithType(errors.TypeNotFound)
			},
			wantErr:     true,
			wantMessage: "stderror",
			wantType:    errors.TypeNotFound,
		},
		{
			name: "multiple levels of wrapping with type and labels, mixed std errors with custom errors",
			makeErr: func() error {
				err := fmt.Errorf("stderror")
				err = errors.From(fmt.Errorf("wrap1: %w", err))
				err = fmt.Errorf("wrap%d: %w", 2, err)
				err = errors.From(err)
				err = errors.From(err).WithType(errors.TypeNotFound)
				err = errors.From(err).WithType(errors.TypeInvalidRequest) // This type should take precedence!
				err = fmt.Errorf("wrap3: %w", err)
				return err
			},
			wantErr:     true,
			wantMessage: "wrap3: wrap2: wrap1: stderror",
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
	key := "testkey"
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
				return errors.New("test").WithValue(key, value)
			},
			wantValue: true,
		},
		{
			name: "nested with value",
			makeErr: func() error {
				return errors.From(
					fmt.Errorf(
						"test: %w",
						errors.New("test").WithValue(key, value),
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

			v := errors.Value(err, key)
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
			defer errors.Finalize(&rerr, "failed to stf")
			return nil
		}()
		assert.True(t, err == nil)
	})
	t.Run("error is wrapped with context", func(t *testing.T) {
		err := func() (rerr error) {
			defer errors.Finalize(&rerr, "foo failed with a=%d", 123)
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
			defer errors.Finalize(&rerr, "func failed")
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
			name: "from",
			makeErr: func() error {
				return errors.From(os.ErrNotExist)
			},
			want: os.ErrNotExist,
		},
		{
			name: "from with context",
			makeErr: func() error {
				return fmt.Errorf("test: %w", errors.From(os.ErrNotExist))
			},
			want: os.ErrNotExist,
		},
		{
			name: "from with context 2",
			makeErr: func() error {
				return errors.From(fmt.Errorf("test: %w", errors.From(os.ErrNotExist)))
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
			name: "from",
			makeErr: func() error {
				return errors.From(&testError{Value: "test"})
			},
		},
		{
			name: "from with context",
			makeErr: func() error {
				return fmt.Errorf("test: %w", errors.From(&testError{Value: "test"}))
			},
		},
		{
			name: "from with context 2",
			makeErr: func() error {
				return errors.From(fmt.Errorf("test: %w", errors.From(&testError{Value: "test"})))
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
