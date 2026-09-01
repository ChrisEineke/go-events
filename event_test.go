package events

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventHasHandlers(t *testing.T) {
	e := E{}
	assert.Equal(t, e.HasHandlers(), false, "there should be no Handlers")

	e.On(func() error { return nil })
	assert.Equal(t, e.HasHandlers(), true, "there should be a Handlers")
}

func TestEventOn(t *testing.T) {
	e := E{}

	err := e.On(func() error { return nil })
	assert.NoError(t, err)

	err = e.On("String")
	assert.Error(t, err)
}

func TestEventOff(t *testing.T) {
	e := E{}
	callable1 := func() error { return nil }
	callable2 := func() error { return nil }

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
	e.On(func(a int) error {
		assert.Equal(t, 10, a)
		return nil
	})
	e.Fire(10)
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

func (t *testware) OnUse(EventSource) error                       { t.onUseCalled++; return nil }
func (t *testware) OnDisuse(EventSource) error                    { t.onDisuseCalled++; return nil }
func (t *testware) OnSubscribe(EventSource, Handler) error        { t.onSubscribeCalled++; return nil }
func (t *testware) OnUnsubscribe(EventSource, Handler) error      { t.onUnsubscribeCalled++; return nil }
func (t *testware) OnAllPreFire(EventSource, ...any) error        { t.onAllPreFireCalled++; return nil }
func (t *testware) OnPreFire(EventSource, Handler, ...any) error  { t.onPreFireCalled++; return nil }
func (t *testware) OnPostFire(EventSource, Handler, ...any) error { t.onPostFireCalled++; return nil }
func (t *testware) OnAllPostFire(EventSource, ...any) error       { t.onAllPostFireCalled++; return nil }

var _ Handlerware = (*testware)(nil)

func TestEventFireWithHandlerware(t *testing.T) {
	e := E{}
	tw := &testware{}
	callable := func() error { return nil }

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
	callable := func() error { return nil }

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
	fn := func() error { flag += 1; return nil }
	e.On(fn, Once())
	e.On(fn)
	e.On(fn)
	e.Fire()

	assert.Equal(t, flag, 3)
}

func TestEventManyOnOnce(t *testing.T) {
	e := E{}
	var flags [3]byte

	e.On(func() error { flags[0]++; return nil }, Once())
	e.On(func() error { flags[1]++; return nil }, Once())
	e.On(func() error { flags[2]++; return nil })

	e.Fire()
	e.Fire()

	assert.Equal(t, flags, [3]byte{1, 1, 2})
}

func TestEventOnOffFunction(t *testing.T) {
	e := E{}
	handler := func() error { return nil }

	e.On(handler)
	err := e.Off(handler)
	assert.NoError(t, err)

	err = e.Off(handler)
	assert.Error(t, err)
}

type testHandler struct {
	val int
}

func (h *testHandler) Handle() error {
	h.val++
	return nil
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
	e.On(func(a int, out *[]int) error {
		*out = append(*out, a)
		return nil
	}, Once(), Async())

	results := []int{}
	e.Fire(10, &results)
	e.Fire(10, &results)
	e.WaitAsync()

	assert.Len(t, results, 1)
	assert.False(t, e.HasHandlers())
}

func TestEventOnAsync(t *testing.T) {
	e := E{}
	e.On(func(a int, out chan<- int) error {
		assert.Equal(t, 1, a)
		out <- a
		close(out)
		return nil
	}, Async())

	results := make(chan int, 1)
	e.Fire(1, results)

	var numResults int64 = 0
	go func() {
		for range results {
			atomic.AddInt64(&numResults, 1)
		}
	}()
	e.WaitAsync()

	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&numResults) == 1
	}, 1*time.Second, 10*time.Millisecond)
}

func TestEventHandlerArgsMismatch(t *testing.T) {
	e := E{}
	e.On(func(a int) error {
		assert.Equal(t, 1, a)
		return nil
	})
	e.Fire(1, 2)
}

func BenchmarkEventFireNoArgs(b *testing.B) {
	e := E{}
	timesCalled := 0
	handler := func() error { timesCalled++; return nil }
	e.On(handler)
	for b.Loop() {
		e.Fire()
	}

	assert.Equal(b, b.N, timesCalled)
}

func BenchmarkEventFireIntArg(b *testing.B) {
	e := E{}
	timesCalled := 0
	handler := func(_ int) error { timesCalled++; return nil }
	e.On(handler)
	for b.Loop() {
		e.Fire(b.N)
	}

	assert.Equal(b, b.N, timesCalled)
}

func BenchmarkEventFireIntIntArg(b *testing.B) {
	e := E{}
	timesCalled := 0
	handler := func(_, _ int) error { timesCalled++; return nil }
	e.On(handler)
	for b.Loop() {
		e.Fire(b.N, b.N)
	}

	assert.Equal(b, b.N, timesCalled)
}
