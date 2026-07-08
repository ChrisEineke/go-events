package events

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvent2HasHandlers(t *testing.T) {
	e := E2[int, int]{}
	assert.Equal(t, e.HasHandlers(), false, "there should be no Handlers")

	e.On(func(arg1, arg2 int) {})
	assert.Equal(t, e.HasHandlers(), true, "there should be a Handlers")
}

func TestEvent2On(t *testing.T) {
	e := E2[int, int]{}

	err := e.On(func(arg1, arg2 int) {})
	assert.NoError(t, err)

	err = e.On("String")
	assert.Error(t, err)
}

func TestEvent2Off(t *testing.T) {
	e := E2[int, int]{}
	callable1 := func(arg1, arg2 int) {}
	callable2 := func(arg1, arg2 int) {}

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

func TestEvent2Fire(t *testing.T) {
	e := E2[int, int]{}
	e.On(func(arg1, arg2 int) {
		assert.Equal(t, 1, arg1)
		assert.Equal(t, 2, arg2)
	})
	e.Fire(1, 2)
}

func TestEvent2Fire2(t *testing.T) {
	e := E2[int, int]{}
	e.On(func(arg1, arg2 int) {
		assert.Equal(t, 1, arg1)
		assert.Equal(t, 2, arg2)
	})
	e.Fire2(1, 2)
}

func TestEvent2Fire2WithHandlerware(t *testing.T) {
	e := E2[int, int]{}
	tw := &testware{}
	callable := func(arg1, arg2 int) {}

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire2(1, 2)
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable)
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire2(1, 2)
	assert.Equal(t, 2, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 1, tw.onPreFireCalled, "OnPreFire should be called for every Handler")
	assert.Equal(t, 1, tw.onPostFireCalled, "OnPostFire should be called for every Handler")
	assert.Equal(t, 2, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.Off(callable)
	assert.Equal(t, 1, tw.onUnsubscribeCalled, "OnUnsubscribe should be called once for every callable detached from the Event")

	e.Disuse(tw)
	assert.Equal(t, 1, tw.onDisuseCalled)
}

func TestEvent2Fire2AsyncWithHandlerware(t *testing.T) {
	e := E2[int, int]{}
	tw := &testware{}
	callable := func(arg1, arg2 int) {}

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire2(1, 2)
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable, Async())
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire2(1, 2)
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

func TestEvent2OnOnceAndManyOn(t *testing.T) {
	e := E2[int, int]{}
	flag := 0
	fn := func(arg1, arg2 int) { flag += 1 }
	e.On(fn, Once())
	e.On(fn)
	e.On(fn)
	e.Fire2(1, 2)

	assert.Equal(t, flag, 3)
}

func TestEvent2ManyOnOnce(t *testing.T) {
	e := E2[int, int]{}
	var flags [3]byte

	e.On(func(arg1, arg2 int) { flags[0]++ }, Once())
	e.On(func(arg1, arg2 int) { flags[1]++ }, Once())
	e.On(func(arg1, arg2 int) { flags[2]++ })

	e.Fire2(1, 2)
	e.Fire2(1, 2)

	assert.Equal(t, flags, [3]byte{1, 1, 2})
}

func TestEvent2OnOffFunction(t *testing.T) {
	e := E2[int, int]{}
	handler := func(arg1, arg2 int) {}

	e.On(handler)
	err := e.Off(handler)
	assert.NoError(t, err)

	err = e.Off(handler)
	assert.Error(t, err)
}

type testHandler2 struct {
	val int
}

func (h *testHandler2) Handle(arg1, arg2 int) {
	h.val++
}

func TestEvent2OnOffReceiver(t *testing.T) {
	e := E2[int, int]{}
	handler := &testHandler2{val: 0}

	e.On(handler.Handle)
	e.Fire2(1, 2)
	err := e.Off(handler.Handle)
	assert.NoError(t, err)

	err = e.Off(handler.Handle)
	assert.Error(t, err)

	e.Fire2(1, 2)
	e.WaitAsync()
	assert.Equal(t, 1, handler.val, "handler wasn't removed after calling Off")
}

func TestEvent2OnOnceAsync(t *testing.T) {
	e := E2[*[]int, *[]int]{}
	e.On(func(out1, out2 *[]int) {
		*out1 = append(*out1, 1)
		*out2 = append(*out2, 2)
	}, Once(), Async())

	results1 := []int{}
	results2 := []int{}
	e.Fire2(&results1, &results2)
	e.Fire2(&results1, &results2)
	e.WaitAsync()

	assert.Len(t, results1, 1)
	assert.Len(t, results2, 1)
	assert.False(t, e.HasHandlers())
}

func TestEvent2OnAsync(t *testing.T) {
	e := E2[chan<- int, chan<- int]{}
	e.On(func(out1, out2 chan<- int) {
		out1 <- 1
		close(out1)
		out2 <- 2
		close(out2)
	}, Async())

	results1 := make(chan int, 1)
	results2 := make(chan int, 1)
	e.Fire2(results1, results2)

	var numResults int64 = 0
	go func() {
		for range results1 {
			atomic.AddInt64(&numResults, 1)
		}
		for range results2 {
			atomic.AddInt64(&numResults, 1)
		}
	}()
	e.WaitAsync()

	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&numResults) == 2
	}, 1*time.Second, 10*time.Millisecond)
}

func BenchmarkEvent2FireIntIntArg(b *testing.B) {
	e := E2[int, int]{}
	timesCalled := 0
	handler := func(_, _ int) { timesCalled++ }
	e.On(handler)
	for b.Loop() {
		e.Fire2(b.N, b.N)
	}

	assert.Equal(b, b.N, timesCalled)
}
