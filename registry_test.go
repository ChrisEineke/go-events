package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryPut(t *testing.T) {
	var err error
	registry := NewRegistry[int]()

	err = registry.Put(nil, 0)
	assert.Error(t, err, "should've failed to register a nil event")

	err = registry.Put(&E{}, 0)
	assert.Error(t, err, "should've failed to register event with blank name")

	err = registry.Put(&E{N: "testEvent"}, 0)
	assert.NoError(t, err, "shouldn't have failed to register a fresh event")

	err = registry.Put(&E{N: "testEvent"}, 1)
	assert.Error(t, err, "should've failed to register event with the same name")
}

func TestRegistryDelete(t *testing.T) {
	var err error
	var data int
	event := &E{N: "testEvent"}
	registry := NewRegistry[int]()

	data, err = registry.Delete(nil)
	assert.Error(t, err, "should've failed to register a nil event")
	assert.Equal(t, 0, data, "data should've been the zero value")

	data, err = registry.Delete(&E{})
	assert.Error(t, err, "should've failed to deregister event with blank name")
	assert.Equal(t, 0, data, "data should've been the zero value")

	err = registry.Put(event, 1)
	assert.NoError(t, err, "shouldn't have failed to register a fresh event")
	data, err = registry.Delete(event)
	assert.NoError(t, err, "shouldn't have failed to deregister event")
	assert.Equal(t, 1, data, "data should've been the value provided at registration")

	data, err = registry.Delete(event)
	assert.Error(t, err, "should've failed to deregister event twice")
	assert.Equal(t, 0, data, "data should've been the zero value")
}

func TestRegistryGet(t *testing.T) {
	var err error
	event := &E{N: "testEvent"}
	registry := NewRegistry[int]()

	err = registry.Put(event, 1)
	assert.NoError(t, err, "shouldn't have failed to register a fresh event")

	registration, err := registry.Get("testEvent")
	assert.NoError(t, err, "shouldn't have failed to get a fresh event")
	assert.NotNil(t, registration, "shouldn't have gotten a nil registration")
	assert.Equal(t, event, registration.event, "should've gotten the same event")
	assert.Equal(t, 1, registration.data, "data should've been the value provided at registration")
}
