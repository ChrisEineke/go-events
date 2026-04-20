package EventBus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventImplementsInterfaces(t *testing.T) {
	var _ Fireable = &Event{}
	var _ Subscribable = &Event{}
	var _ Waitable = &Event{}
}

func TestEventHasListeners(t *testing.T) {
	e := &Event{}
	assert.Equal(t, e.HasListeners(), false, "there should be no listeners")

	e.On(func() {})
	assert.Equal(t, e.HasListeners(), true, "there should be a listeners")
}

func TestEventOn(t *testing.T) {
	e := &Event{}

	err := e.On(func() {})
	assert.NoError(t, err)

	err = e.On("String")
	assert.Error(t, err)
}

func TestEventOnOnceAndManyOn(t *testing.T) {
	e := &Event{}
	flag := 0
	fn := func() { flag += 1 }
	e.On(fn, Once())
	e.On(fn)
	e.On(fn)
	e.Fire()

	assert.Equal(t, flag, 3)
}

func TestEventManyOnOnce(t *testing.T) {
	e := &Event{}
	var flags [3]byte

	e.On(func() { flags[0]++ }, Once())
	e.On(func() { flags[1]++ }, Once())
	e.On(func() { flags[2]++ })

	e.Fire()
	e.Fire()

	assert.Equal(t, flags, [3]byte{1, 1, 2})
}

func TestEventOnOffFunction(t *testing.T) {
	e := &Event{}
	handler := func() {}

	e.On(handler)
	err := e.Off(handler)
	assert.NoError(t, err)

	err = e.Off(handler)
	assert.Error(t, err)
}

type handler struct {
	val int
}

func (h *handler) Handle() {
	h.val++
}

func TestEventOnOffReceiver(t *testing.T) {
	e := &Event{}
	handler := &handler{val: 0}

	e.On(handler.Handle)
	e.Fire()
	err := e.Off(handler.Handle)
	assert.NoError(t, err)

	err = e.Off(handler.Handle)
	assert.Error(t, err)

	e.Fire()
	e.WaitAsync()
	assert.Equal(t, 1, handler.val, "handler wasn't removed after calling Off")
}

func TestEventFire(t *testing.T) {
	e := &Event{}
	e.On(func(a int, err error) {
		assert.Equal(t, 10, a)
		assert.NoError(t, err)
	})
	e.Fire(10, nil)
}

func TestEventOnOnceAsync(t *testing.T) {
	e := &Event{}
	e.On(func(a int, out *[]int) {
		*out = append(*out, a)
	}, Once(), Async())

	results := []int{}
	e.Fire(10, &results)
	e.Fire(10, &results)
	e.WaitAsync()

	assert.Len(t, results, 1)
	assert.False(t, e.HasListeners())
}

func TestEventOnAsyncTransactional(t *testing.T) {
	e := &Event{}
	e.On(func(a int, out *[]int, dur string) {
		sleep, _ := time.ParseDuration(dur)
		time.Sleep(sleep)
		*out = append(*out, a)
	}, Async(), Transactional())

	results := make([]int, 0)
	e.Fire(1, &results, "1s")
	e.Fire(2, &results, "0s")
	e.WaitAsync()

	assert.Len(t, results, 2)
	assert.Equal(t, 1, results[0])
	assert.Equal(t, 2, results[1])
}

func TestEventOnAsync(t *testing.T) {
	e := &Event{}
	e.On(func(a int, out chan<- int) {
		out <- a
	}, Async())

	results := make(chan int)
	e.Fire(1, results)
	e.Fire(2, results)

	numResults := 0
	go func() {
		for range results {
			numResults++
		}
	}()
	e.WaitAsync()

	assert.Eventually(t, func() bool { return numResults == 2 }, 1*time.Second, 10*time.Millisecond)
}

func TestEventListenerArgsMismatch(t *testing.T) {
	e := &Event{}
	e.On(func(a int) {
		assert.Equal(t, 1, a)
	})
	e.Fire(1, 2)
}

func BenchmarkEventFireNoArgs(b *testing.B) {
	e := &Event{}
	timesCalled := 0
	handler := func() { timesCalled++ }
	e.On(handler)
	for b.Loop() {
		e.Fire()
	}

	assert.Equal(b, b.N, timesCalled)
}

func BenchmarkEventFireIntArg(b *testing.B) {
	e := &Event{}
	timesCalled := 0
	handler := func(_ int) { timesCalled++ }
	e.On(handler)
	for b.Loop() {
		e.Fire(b.N)
	}

	assert.Equal(b, b.N, timesCalled)
}

func BenchmarkEventFireIntIntArg(b *testing.B) {
	e := &Event{}
	timesCalled := 0
	handler := func(_, _ int) { timesCalled++ }
	e.On(handler)
	for b.Loop() {
		e.Fire(b.N, b.N)
	}

	assert.Equal(b, b.N, timesCalled)
}
