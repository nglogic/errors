package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_UnknownTypeDefaults(t *testing.T) {
	require.Equal(t, "Unknown", TypeUnknown.String())
	require.Equal(t, GroupServer, TypeUnknown.Group)
}

func TestGetType(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Type
	}{
		// {
		// 	name: "nil",
		// 	err:  nil,
		// 	want: TypeUnknown,
		// },
		// {
		// 	name: "no type",
		// 	err:  New("err"),
		// 	want: TypeUnknown,
		// },
		// {
		// 	name: "with type",
		// 	err:  New("err").WithType(TypeInvalidRequest),
		// 	want: TypeInvalidRequest,
		// },
		// {
		// 	name: "from fmt.Errorf wrap with type",
		// 	err:  fmt.Errorf("wrap: %w", FromErr(stderrors.New("err")).WithType(TypeInvalidRequest)),
		// 	want: TypeInvalidRequest,
		// },
		// {
		// 	name: "from internal err with type",
		// 	err:  FromErr(New("err")).WithType(TypeInvalidRequest),
		// 	want: TypeInvalidRequest,
		// },
		{
			name: "from internal err with type v2",
			err:  FromErr(New("err").WithType(TypeInvalidRequest)),
			want: TypeInvalidRequest,
		},
		// {
		// 	name: "multiple types, want most recent one",
		// 	err: FromErr(
		// 		New("err").WithType(TypeInvalidRequest),
		// 	).WithType(TypeAlreadyExists),
		// 	want: TypeAlreadyExists,
		// },
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tp := GetType(tt.err)
			t.Log(tp.String())
			assert.Equal(t, tt.want, tp)
			assert.True(t, IsType(tt.err, tp))
			assert.True(t, IsGroup(tt.err, tp.Group))

			grp := GetGroup(tt.err)
			t.Log(grp.String())
			assert.Equal(t, tt.want.Group, grp)
		})
	}
}
