package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventImplementsInterfaces(t *testing.T) {
	var _ Fireable = &E{}
	var _ Subscribable = &E{}
	var _ Waitable = &E{}
}

func TestEventHasHandlers(t *testing.T) {
	e := E{}
	assert.Equal(t, e.HasHandlers(), false, "there should be no Handlers")

	e.On(func() {})
	assert.Equal(t, e.HasHandlers(), true, "there should be a Handlers")
}

func TestEventOn(t *testing.T) {
	e := E{}

	err := e.On(func() {})
	assert.NoError(t, err)

	err = e.On("String")
	assert.Error(t, err)
}

func TestEventOff(t *testing.T) {
	e := E{}
	callable1 := func() {}
	callable2 := func() {}

	err := e.On(callable1)
	assert.NoError(t, err)
	err = e.On(callable2)
	assert.NoError(t, err)
	err = e.On(callable2)
	assert.NoError(t, err)

	err = e.Off(callable1)
	assert.NoError(t, err)
	err = e.Off(callable2)
	assert.NoError(t, err)
	err = e.Off(callable2)
	assert.NoError(t, err)

	err = e.Off(callable1)
	assert.Error(t, err)
	err = e.Off(callable2)
	assert.Error(t, err)
}

func TestEventFire(t *testing.T) {
	e := E{}
	e.On(func(a int, err error) {
		assert.Equal(t, 10, a)
		assert.NoError(t, err)
	})
	e.Fire(10, nil)
}

type testware struct {
	Handlerware

	onUseCalled         int
	onDisuseCalled      int
	onSubscribeCalled   int
	onUnsubscribeCalled int
	onAllPreFireCalled  int
	onPreFireCalled     int
	onPostFireCalled    int
	onAllPostFireCalled int
}

func (t *testware) OnUse(*E) error                { t.onUseCalled++; return nil }
func (t *testware) OnDisuse(*E) error             { t.onDisuseCalled++; return nil }
func (t *testware) OnSubscribe(*E, Handler)       { t.onSubscribeCalled++ }
func (t *testware) OnUnsubscribe(*E, Handler)     { t.onUnsubscribeCalled++ }
func (t *testware) OnAllPreFire(*E, []any)        { t.onAllPreFireCalled++ }
func (t *testware) OnPreFire(*E, Handler, []any)  { t.onPreFireCalled++ }
func (t *testware) OnPostFire(*E, Handler, []any) { t.onPostFireCalled++ }
func (t *testware) OnAllPostFire(*E, []any)       { t.onAllPostFireCalled++ }

func TestEventFireWithHandlerware(t *testing.T) {
	e := E{}
	tw := &testware{}
	callable := func() {}

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire()
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable)
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire()
	assert.Equal(t, 2, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 1, tw.onPreFireCalled, "OnPreFire should be called for every Handler")
	assert.Equal(t, 1, tw.onPostFireCalled, "OnPostFire should be called for every Handler")
	assert.Equal(t, 2, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.Off(callable)
	assert.Equal(t, 1, tw.onUnsubscribeCalled, "OnUnsubscribe should be called once for every callable detached from the Event")

	e.Disuse(tw)
	assert.Equal(t, 1, tw.onDisuseCalled)
}

func TestEventFireAsyncWithHandlerware(t *testing.T) {
	e := E{}
	tw := &testware{}
	callable := func() {}

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire()
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable, Async())
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire()
	e.WaitAsync()
	assert.Equal(t, 2, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 1, tw.onPreFireCalled, "OnPreFire should be called for every Handler")
	assert.Equal(t, 1, tw.onPostFireCalled, "OnPostFire should be called for every Handler")
	assert.Equal(t, 2, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.Off(callable)
	assert.Equal(t, 1, tw.onUnsubscribeCalled, "OnUnsubscribe should be called once for every callable detached from the Event")

	e.Disuse(tw)
	assert.Equal(t, 1, tw.onDisuseCalled)
}

func TestEventOnOnceAndManyOn(t *testing.T) {
	e := E{}
	flag := 0
	fn := func() { flag += 1 }
	e.On(fn, Once())
	e.On(fn)
	e.On(fn)
	e.Fire()

	assert.Equal(t, flag, 3)
}

func TestEventManyOnOnce(t *testing.T) {
	e := E{}
	var flags [3]byte

	e.On(func() { flags[0]++ }, Once())
	e.On(func() { flags[1]++ }, Once())
	e.On(func() { flags[2]++ })

	e.Fire()
	e.Fire()

	assert.Equal(t, flags, [3]byte{1, 1, 2})
}

func TestEventOnOffFunction(t *testing.T) {
	e := E{}
	handler := func() {}

	e.On(handler)
	err := e.Off(handler)
	assert.NoError(t, err)

	err = e.Off(handler)
	assert.Error(t, err)
}

type testHandler struct {
	val int
}

func (h *testHandler) Handle() {
	h.val++
}

func TestEventOnOffReceiver(t *testing.T) {
	e := E{}
	handler := &testHandler{val: 0}

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

func TestEventOnOnceAsync(t *testing.T) {
	e := E{}
	e.On(func(a int, out *[]int) {
		*out = append(*out, a)
	}, Once(), Async())

	results := []int{}
	e.Fire(10, &results)
	e.Fire(10, &results)
	e.WaitAsync()

	assert.Len(t, results, 1)
	assert.False(t, e.HasHandlers())
}

func TestEventOnAsyncTransactional(t *testing.T) {
	e := E{}
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
	e := E{}
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

func TestEventHandlerArgsMismatch(t *testing.T) {
	e := E{}
	e.On(func(a int) {
		assert.Equal(t, 1, a)
	})
	e.Fire(1, 2)
}

func BenchmarkEventFireNoArgs(b *testing.B) {
	e := E{}
	timesCalled := 0
	handler := func() { timesCalled++ }
	e.On(handler)
	for b.Loop() {
		e.Fire()
	}

	assert.Equal(b, b.N, timesCalled)
}

func BenchmarkEventFireIntArg(b *testing.B) {
	e := E{}
	timesCalled := 0
	handler := func(_ int) { timesCalled++ }
	e.On(handler)
	for b.Loop() {
		e.Fire(b.N)
	}

	assert.Equal(b, b.N, timesCalled)
}

func BenchmarkEventFireIntIntArg(b *testing.B) {
	e := E{}
	timesCalled := 0
	handler := func(_, _ int) { timesCalled++ }
	e.On(handler)
	for b.Loop() {
		e.Fire(b.N, b.N)
	}

	assert.Equal(b, b.N, timesCalled)
}
