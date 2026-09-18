package events

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler3(t *testing.T) {
	fn := func(arg1, arg2, arg3 int) error {
		assert.Equal(t, 1, arg1)
		assert.Equal(t, 2, arg2)
		assert.Equal(t, 3, arg3)
		return nil
	}

	h, err := newHandler3[int, int, int](&E3[int, int, int]{}, fn)
	assert.NoError(t, err)
	assert.IsType(t, &handler3[int, int, int]{}, h)
	assert.Equal(t, reflect.ValueOf(fn), h.funcValue())

	h.apply3(1, 2, 3)
	h.apply(1, 2, 3)

	assert.Error(t, h.apply(), "expected error; requires exactly 3 argument")
}
