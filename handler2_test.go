package events

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler2(t *testing.T) {
	callback := func(arg1, arg2 int) error {
		assert.Equal(t, 1, arg1)
		assert.Equal(t, 2, arg2)
		return nil
	}

	h, err := newHandler2[int, int](&E2[int, int]{}, callback)
	assert.NoError(t, err)
	assert.IsType(t, &handler2[int, int]{}, h)
	assert.Equal(t, reflect.ValueOf(callback).Pointer(), h.callable().Pointer())

	h.apply2(1, 2)
	h.apply(1, 2)

	assert.Error(t, h.apply(), "expected error; requires exactly 2 arguments")
}
