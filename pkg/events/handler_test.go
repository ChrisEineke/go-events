package events

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullaryHandler(t *testing.T) {
	fn := func() error { return nil }

	h, err := newHandler(&E{}, fn)
	assert.NoError(t, err)
	assert.IsType(t, &nullaryHandler{}, h)

	assert.Equal(t, reflect.ValueOf(fn), h.funcValue())
}

func TestNAryHandler(t *testing.T) {
	fn1 := func(a int) error { return nil }
	fn2 := func(a, b int) error { return nil }

	h1, err := newHandler(&E{}, fn1)
	assert.NoError(t, err)
	assert.IsType(t, &nAryHandler{}, h1)

	assert.Equal(t, reflect.ValueOf(fn1), h1.funcValue())

	h2, err := newHandler(&E{}, fn2)
	assert.NoError(t, err)
	assert.IsType(t, &nAryHandler{}, h2)

	assert.Equal(t, reflect.ValueOf(fn2), h2.funcValue())
}
