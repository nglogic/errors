package errors

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UnknownType(t *testing.T) {
	require.Equal(t, "Unknown", TypeUnknown.String())
}

func TestGetType(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Type
	}{
		{
			name: "nil",
			err:  nil,
			want: TypeUnknown,
		},
		{
			name: "no type",
			err:  New("err"),
			want: TypeUnknown,
		},
		{
			name: "with type",
			err:  New("err").WithType(TypeInvalidRequest),
			want: TypeInvalidRequest,
		},
		{
			name: "from fmt.Errorf wrap with type",
			err:  fmt.Errorf("wrap: %w", From(stderrors.New("err")).WithType(TypeInvalidRequest)),
			want: TypeInvalidRequest,
		},
		{
			name: "from internal err with type",
			err:  From(New("err")).WithType(TypeInvalidRequest),
			want: TypeInvalidRequest,
		},
		{
			name: "from internal err with type v2",
			err:  From(New("err").WithType(TypeInvalidRequest)),
			want: TypeInvalidRequest,
		},
		{
			name: "multiple types, want most recent one",
			err: From(
				New("err").WithType(TypeInvalidRequest),
			).WithType(TypeAlreadyExists),
			want: TypeAlreadyExists,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tp := GetType(tt.err)
			t.Log(tp.String())
			assert.Equal(t, tt.want, tp)
			assert.True(t, IsType(tt.err, tp))
		})
	}
}
