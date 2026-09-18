package events

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler1(t *testing.T) {
	fn := func(arg1 int) error {
		assert.Equal(t, 1, arg1)
		return nil
	}

	h, err := newHandler1[int](&E1[int]{}, fn)
	assert.NoError(t, err)
	assert.IsType(t, &handler1[int]{}, h)
	assert.Equal(t, reflect.ValueOf(fn), h.funcValue())

	h.apply1(1)

	assert.Error(t, h.apply(), "expected error; requires exactly 1 argument")
}
