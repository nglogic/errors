package errors

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	t.Run("error is nil", func(t *testing.T) {
		err := func() (rerr error) {
			defer Wrap(&rerr, "failed to stf")
			return nil
		}()
		assert.NoError(t, err)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		err := func() (rerr error) {
			defer Wrap(&rerr, "foo failed with a=%d", 123)
			return io.ErrClosedPipe
		}()
		assert.Error(t, err)
		assert.Equal(t,
			fmt.Sprintf("foo failed with a=123: %s", io.ErrClosedPipe.Error()),
			err.Error(),
		)
	})

	t.Run("GRPC error code can be unwrapped", func(t *testing.T) {
		err := func() (rerr error) {
			defer Wrap(&rerr, "func failed")
			return New("not found").WithType(TypeNotFound)
		}()
		assert.Error(t, err)
		assert.True(t, IsType(err, TypeNotFound))
	})
}
