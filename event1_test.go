package events

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvent1HasHandlers(t *testing.T) {
	e := E1[int]{}
	assert.Equal(t, e.HasHandlers(), false, "there should be no Handlers")

	e.On(func(arg1 int) {})
	assert.Equal(t, e.HasHandlers(), true, "there should be a Handlers")
}

func TestEvent1On(t *testing.T) {
	e := E1[int]{}

	err := e.On(func(arg1 int) {})
	assert.NoError(t, err)

	err = e.On("String")
	assert.Error(t, err)
}

func TestEvent1Off(t *testing.T) {
	e := E1[int]{}
	callable1 := func(arg1 int) {}
	callable2 := func(arg1 int) {}

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

func TestEvent1Fire(t *testing.T) {
	e := E1[int]{}
	e.On(func(arg1 int) {
		assert.Equal(t, 1, arg1)
	})
	e.Fire(1)
}

func TestEvent1Fire1(t *testing.T) {
	e := E1[int]{}
	e.On(func(a int) {
		assert.Equal(t, 1, a)
	})
	e.Fire1(1)
}

func TestEvent1Fire1WithHandlerware(t *testing.T) {
	e := E1[int]{}
	tw := &testware{}
	callable := func(arg1 int) {}

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire1(1)
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable)
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire1(1)
	assert.Equal(t, 2, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 1, tw.onPreFireCalled, "OnPreFire should be called for every Handler")
	assert.Equal(t, 1, tw.onPostFireCalled, "OnPostFire should be called for every Handler")
	assert.Equal(t, 2, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.Off(callable)
	assert.Equal(t, 1, tw.onUnsubscribeCalled, "OnUnsubscribe should be called once for every callable detached from the Event")

	e.Disuse(tw)
	assert.Equal(t, 1, tw.onDisuseCalled)
}

func TestEvent1Fire1AsyncWithHandlerware(t *testing.T) {
	e := E1[int]{}
	tw := &testware{}
	callable := func(arg1 int) {}

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire1(1)
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable, Async())
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire1(1)
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

func TestEvent1OnOnceAndManyOn(t *testing.T) {
	e := E1[int]{}
	flag := 0
	fn := func(arg1 int) { flag += 1 }
	e.On(fn, Once())
	e.On(fn)
	e.On(fn)
	e.Fire1(1)

	assert.Equal(t, flag, 3)
}

func TestEvent1ManyOnOnce(t *testing.T) {
	e := E1[int]{}
	var flags [3]byte

	e.On(func(arg1 int) { flags[0]++ }, Once())
	e.On(func(arg1 int) { flags[1]++ }, Once())
	e.On(func(arg1 int) { flags[2]++ })

	e.Fire1(1)
	e.Fire1(1)

	assert.Equal(t, flags, [3]byte{1, 1, 2})
}

func TestEvent1OnOffFunction(t *testing.T) {
	e := E1[int]{}
	handler := func(arg1 int) {}

	e.On(handler)
	err := e.Off(handler)
	assert.NoError(t, err)

	err = e.Off(handler)
	assert.Error(t, err)
}

type testHandler1 struct {
	val int
}

func (h *testHandler1) Handle(arg1 int) {
	h.val++
}

func TestEvent1OnOffReceiver(t *testing.T) {
	e := E1[int]{}
	handler := &testHandler1{val: 0}

	e.On(handler.Handle)
	e.Fire1(1)
	err := e.Off(handler.Handle)
	assert.NoError(t, err)

	err = e.Off(handler.Handle)
	assert.Error(t, err)

	e.Fire1(1)
	e.WaitAsync()
	assert.Equal(t, 1, handler.val, "handler wasn't removed after calling Off")
}

func TestEvent1OnOnceAsync(t *testing.T) {
	e := E1[*[]int]{}
	e.On(func(out *[]int) {
		*out = append(*out, 1)
	}, Once(), Async())

	results := []int{}
	e.Fire1(&results)
	e.Fire1(&results)
	e.WaitAsync()

	assert.Len(t, results, 1)
	assert.False(t, e.HasHandlers())
}

func TestEvent1OnAsync(t *testing.T) {
	e := E1[chan<- int]{}
	e.On(func(out chan<- int) {
		out <- 1
		close(out)
	}, Async())

	results := make(chan int, 2)
	e.Fire1(results)

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

func BenchmarkEvent1FireIntArg(b *testing.B) {
	e := E1[int]{}
	timesCalled := 0
	handler := func(_ int) { timesCalled++ }
	e.On(handler)
	for b.Loop() {
		e.Fire1(b.N)
	}

	assert.Equal(b, b.N, timesCalled)
}
