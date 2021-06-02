package errors_test

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/nglogic/errors"
	"github.com/stretchr/testify/assert"
)

func TestSimpleError(t *testing.T) {
	tests := []struct {
		name        string
		makeErr     func() error
		wantMessage string
		wantType    errors.Type
		wantGroup   errors.Group
	}{
		{
			name: "nil of type errors.Error",
			makeErr: func() error {
				var err errors.Error
				return err
			},
			wantMessage: "",
			wantType:    errors.TypeUnknown,
			wantGroup:   errors.GroupServer,
		},
		{
			name: "simple",
			makeErr: func() error {
				return errors.New("test")
			},
			wantMessage: "test",
			wantType:    errors.TypeUnknown,
			wantGroup:   errors.GroupServer,
		},
		{
			name: "simple with formatting",
			makeErr: func() error {
				return errors.Newf("test %d %d %d", 1, 2, 3)
			},
			wantMessage: "test 1 2 3",
			wantType:    errors.TypeUnknown,
			wantGroup:   errors.GroupServer,
		},
		{
			name: "simple with type",
			makeErr: func() error {
				return errors.New("test").WithType(errors.TypeNotFound)
			},
			wantMessage: "test",
			wantType:    errors.TypeNotFound,
			wantGroup:   errors.GroupClient,
		},
		{
			name: "wapped with type",
			makeErr: func() error {
				err := stderrors.New("stderror")
				return errors.FromErr(err).WithType(errors.TypeNotFound)
			},
			wantMessage: "stderror",
			wantType:    errors.TypeNotFound,
			wantGroup:   errors.GroupClient,
		},
		{
			name: "multiple levels of wrapping with type and labels, mixed std errors with custom errors",
			makeErr: func() error {
				err := stderrors.New("stderror")
				err = errors.FromErr(fmt.Errorf("wrap1: %w", err))
				err = fmt.Errorf("wrap%d: %w", 2, err)
				err = errors.FromErr(err)
				err = errors.FromErr(err).WithType(errors.TypeNotFound)
				err = errors.FromErr(err).WithType(errors.TypeInvalidRequest) // This type should take precedence!
				err = fmt.Errorf("wrap3: %w", err)
				return err
			},
			wantMessage: "wrap3: wrap2: wrap1: stderror",
			wantType:    errors.TypeInvalidRequest,
			wantGroup:   errors.GroupClient,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.makeErr()

			if err != nil {
				assert.Equal(t, tt.wantMessage, err.Error(), "invalid error message")
			}

			assert.True(t, errors.IsType(err, tt.wantType), "invalid error type chek result")
			assert.Equal(t, tt.wantType, errors.GetType(err), "invalid error type")

			assert.True(t, errors.IsGroup(err, tt.wantGroup), "invalid error group check result")
			assert.Equal(t, tt.wantGroup, errors.GetGroup(err), "invalid error group")
		})
	}
}

func TestNilError(t *testing.T) {
	err := errors.FromErr(nil)
	assert.True(t, err == nil)
}
