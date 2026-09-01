package events

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler4(t *testing.T) {
	callback := func(arg1, arg2, arg3, arg4 int) error {
		assert.Equal(t, 1, arg1)
		assert.Equal(t, 2, arg2)
		assert.Equal(t, 3, arg3)
		assert.Equal(t, 4, arg4)
		return nil
	}

	h, err := newHandler4[int, int, int, int](&E4[int, int, int, int]{}, callback)
	assert.NoError(t, err)
	assert.IsType(t, &handler4[int, int, int, int]{}, h)
	assert.Equal(t, reflect.ValueOf(callback).Pointer(), h.callable().Pointer())

	h.apply4(1, 2, 3, 4)
	h.apply(1, 2, 3, 4)

	assert.Error(t, h.apply(), "expected error; requires exactly 4 argument")
}
