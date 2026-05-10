package events

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullaryHander(t *testing.T) {
	callback := func() {}

	h, err := newHandler(callback)
	assert.NoError(t, err)
	assert.IsType(t, &nullaryHandler{}, h)

	h.Lock()
	assert.Equal(t, reflect.ValueOf(callback).Pointer(), h.getCallable().Pointer())
	h.Unlock()
}

func TestNAryHandler(t *testing.T) {
	callback1 := func(a int) {}
	callback2 := func(a, b int) {}

	h1, err := newHandler(callback1)
	assert.NoError(t, err)
	assert.IsType(t, &nAryHandler{}, h1)

	h1.Lock()
	assert.Equal(t, reflect.ValueOf(callback1).Pointer(), h1.getCallable().Pointer())
	h1.Unlock()

	h2, err := newHandler(callback2)
	assert.NoError(t, err)
	assert.IsType(t, &nAryHandler{}, h2)

	h2.Lock()
	assert.Equal(t, reflect.ValueOf(callback2).Pointer(), h2.getCallable().Pointer())
	h2.Unlock()
}
