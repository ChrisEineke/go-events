package events

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler2(t *testing.T) {
	callback := func(arg1, arg2 int) {
		assert.Equal(t, 1, arg1)
		assert.Equal(t, 2, arg2)
	}

	h, err := newHandler2(callback)
	assert.NoError(t, err)
	assert.IsType(t, &handler2[int, int]{}, h)
	assert.Equal(t, reflect.ValueOf(callback).Pointer(), h.getCallable().Pointer())

	h.apply2(1, 2)
	h.apply(1, 2)

	assert.Error(t, h.apply(), "expected error; requires exactly 1 argument")
}
