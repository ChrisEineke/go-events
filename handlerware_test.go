package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryRegister(t *testing.T) {
	var err error
	registry := NewRegistry[int]()

	err = registry.Register(nil, 0)
	assert.Error(t, err, "should've failed to register a nil event")

	err = registry.Register(&Event{}, 0)
	assert.Error(t, err, "should've failed to register event with blank name")

	err = registry.Register(&Event{N: "testEvent"}, 0)
	assert.NoError(t, err, "shouldn't have failed to register a fresh event")

	err = registry.Register(&Event{N: "testEvent"}, 1)
	assert.Error(t, err, "should've failed to register event with the same name")
}

func TestRegistryDeregister(t *testing.T) {
	var err error
	var data int
	event := &Event{N: "testEvent"}
	registry := NewRegistry[int]()

	data, err = registry.Deregister(nil)
	assert.Error(t, err, "should've failed to register a nil event")
	assert.Equal(t, 0, data, "data should've been the zero value")

	data, err = registry.Deregister(&Event{})
	assert.Error(t, err, "should've failed to deregister event with blank name")
	assert.Equal(t, 0, data, "data should've been the zero value")

	err = registry.Register(event, 1)
	assert.NoError(t, err, "shouldn't have failed to register a fresh event")
	data, err = registry.Deregister(event)
	assert.NoError(t, err, "shouldn't have failed to deregister event")
	assert.Equal(t, 1, data, "data should've been the value provided at registration")

	data, err = registry.Deregister(event)
	assert.Error(t, err, "should've failed to deregister event twice")
	assert.Equal(t, 0, data, "data should've been the zero value")
}

func TestRegistryGet(t *testing.T) {
	var err error
	event := &Event{N: "testEvent"}
	registry := NewRegistry[int]()

	err = registry.Register(event, 1)
	assert.NoError(t, err, "shouldn't have failed to register a fresh event")

	registration, err := registry.Get("testEvent")
	assert.NoError(t, err, "shouldn't have failed to get a fresh event")
	assert.NotNil(t, registration, "shouldn't have gotten a nil registration")
	assert.Equal(t, event, registration.event, "should've gotten the same event")
	assert.Equal(t, 1, registration.data, "data should've been the value provided at registration")
}
