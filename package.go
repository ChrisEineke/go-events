package EventBus

import "sync"

var (
	singleton Bus
	once      sync.Once
)

// Singleton returns a shared Bus instance. This instance may have been modified by other code to contain topics and
// subscriptions.
func Singleton() Bus {
	once.Do(func() {
		singleton = New()
	})
	return singleton
}

// New returns a new Bus instance with no topics and no subscriptions.
func New() Bus {
	b := &bus{
		registry:     make(map[string]Topic, 1),
		registryLock: sync.Mutex{},
		wg:           sync.WaitGroup{},
	}
	return Bus(b)
}
