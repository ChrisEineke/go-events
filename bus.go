package EventBus

import (
	"fmt"
	"sync"
)

// Bus is the root interface for managing topics (safely and unsafely).
type Bus interface {
	// Topic returns the topic with the given name. If a topic doesn't exist, it will create one.
	Topic(name string) Topic
	// SafeTopic returns the topic with the given name. It will return an error if the topic doesn't exist.
	SafeTopic(name string) (Topic, error)
	// WaitAsync waits for any and all async listeners to complete.
	WaitAsync()
}

// bus is the general implementation of the Bus interface.
type bus struct {
	registry     map[string]Topic
	registryLock sync.Mutex
	wg           sync.WaitGroup
}

func (b *bus) Topic(name string) Topic {
	b.registryLock.Lock()
	defer b.registryLock.Unlock()

	t, ok := b.registry[name]
	if !ok {
		t = newTopic(name, &b.wg)
		b.registry[name] = t
	}
	return t
}

func (b *bus) SafeTopic(name string) (Topic, error) {
	b.registryLock.Lock()
	defer b.registryLock.Unlock()

	t, ok := b.registry[name]
	if !ok {
		return nil, fmt.Errorf("topic %s doesn't exist", name)
	}
	return t, nil
}

// WaitAsync waits for all async listeners to complete
func (bus *bus) WaitAsync() {
	bus.wg.Wait()
}
