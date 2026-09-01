package events

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullaryHandler(t *testing.T) {
	callback := func() error { return nil }

	h, err := newHandler(&E{}, callback)
	assert.NoError(t, err)
	assert.IsType(t, &nullaryHandler{}, h)

	assert.Equal(t, reflect.ValueOf(callback).Pointer(), h.callable().Pointer())
}

func TestNAryHandler(t *testing.T) {
	callback1 := func(a int) error { return nil }
	callback2 := func(a, b int) error { return nil }

	h1, err := newHandler(&E{}, callback1)
	assert.NoError(t, err)
	assert.IsType(t, &nAryHandler{}, h1)

	assert.Equal(t, reflect.ValueOf(callback1).Pointer(), h1.callable().Pointer())

	h2, err := newHandler(&E{}, callback2)
	assert.NoError(t, err)
	assert.IsType(t, &nAryHandler{}, h2)

	assert.Equal(t, reflect.ValueOf(callback2).Pointer(), h2.callable().Pointer())
}
