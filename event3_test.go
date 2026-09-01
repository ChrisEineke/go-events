package events

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvent3HasHandlers(t *testing.T) {
	e := E3[int, int, int]{}
	assert.Equal(t, e.HasHandlers(), false, "there should be no Handlers")

	e.On(func(arg1, arg2, arg3 int) error { return nil })
	assert.Equal(t, e.HasHandlers(), true, "there should be a Handlers")
}

func TestEvent3On(t *testing.T) {
	e := E3[int, int, int]{}

	err := e.On(func(arg1, arg2, arg3 int) error { return nil })
	assert.NoError(t, err)
}

func TestEvent3Off(t *testing.T) {
	e := E3[int, int, int]{}
	callable1 := func(arg1, arg2, arg3 int) error { return nil }
	callable2 := func(arg1, arg2, arg3 int) error { return nil }

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

func TestEvent3Fire(t *testing.T) {
	e := E3[int, int, int]{}
	e.On(func(arg1, arg2, arg3 int) error {
		assert.Equal(t, 1, arg1)
		assert.Equal(t, 2, arg2)
		assert.Equal(t, 3, arg3)
		return nil
	})
	e.Fire(1, 2, 3)
}

func TestEvent3Fire3(t *testing.T) {
	e := E3[int, int, int]{}
	e.On(func(arg1, arg2, arg3 int) error {
		assert.Equal(t, 1, arg1)
		assert.Equal(t, 2, arg2)
		assert.Equal(t, 3, arg3)
		return nil
	})
	e.Fire3(1, 2, 3)
}

func TestEvent3Fire3WithHandlerware(t *testing.T) {
	e := E3[int, int, int]{}
	tw := &testware{}
	callable := func(arg1, arg2, arg3 int) error { return nil }

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire3(1, 2, 3)
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable)
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire3(1, 2, 3)
	assert.Equal(t, 2, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 1, tw.onPreFireCalled, "OnPreFire should be called for every Handler")
	assert.Equal(t, 1, tw.onPostFireCalled, "OnPostFire should be called for every Handler")
	assert.Equal(t, 2, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.Off(callable)
	assert.Equal(t, 1, tw.onUnsubscribeCalled, "OnUnsubscribe should be called once for every callable detached from the Event")

	e.Disuse(tw)
	assert.Equal(t, 1, tw.onDisuseCalled)
}

func TestEvent3Fire3AsyncWithHandlerware(t *testing.T) {
	e := E3[int, int, int]{}
	tw := &testware{}
	callable := func(arg1, arg2, arg3 int) error { return nil }

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire3(1, 2, 3)
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable, Async())
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire3(1, 2, 3)
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

func TestEvent3OnOnceAndManyOn(t *testing.T) {
	e := E3[int, int, int]{}
	flag := 0
	fn := func(arg1, arg2, arg3 int) error { flag += 1; return nil }
	e.On(fn, Once())
	e.On(fn)
	e.On(fn)
	e.Fire3(1, 2, 3)

	assert.Equal(t, flag, 3)
}

func TestEvent3ManyOnOnce(t *testing.T) {
	e := E3[int, int, int]{}
	var flags [3]byte

	e.On(func(arg1, arg2, arg3 int) error { flags[0]++; return nil }, Once())
	e.On(func(arg1, arg2, arg3 int) error { flags[1]++; return nil }, Once())
	e.On(func(arg1, arg2, arg3 int) error { flags[2]++; return nil })

	e.Fire3(1, 2, 3)
	e.Fire3(1, 2, 3)

	assert.Equal(t, flags, [3]byte{1, 1, 2})
}

func TestEvent3OnOffFunction(t *testing.T) {
	e := E3[int, int, int]{}
	handler := func(arg1, arg2, arg3 int) error { return nil }

	e.On(handler)
	err := e.Off(handler)
	assert.NoError(t, err)

	err = e.Off(handler)
	assert.Error(t, err)
}

type testHandler3 struct {
	val int
}

func (h *testHandler3) Handle(arg1, arg2, arg3 int) error {
	h.val++
	return nil
}

func TestEvent3OnOffReceiver(t *testing.T) {
	e := E3[int, int, int]{}
	handler := &testHandler3{val: 0}

	e.On(handler.Handle)
	e.Fire3(1, 2, 3)
	err := e.Off(handler.Handle)
	assert.NoError(t, err)

	err = e.Off(handler.Handle)
	assert.Error(t, err)

	e.Fire3(1, 2, 3)
	e.WaitAsync()
	assert.Equal(t, 1, handler.val, "handler wasn't removed after calling Off")
}

func TestEvent3OnOnceAsync(t *testing.T) {
	e := E3[*[]int, *[]int, *[]int]{}
	e.On(func(out1, out2, out3 *[]int) error {
		*out1 = append(*out1, 1)
		*out2 = append(*out2, 2)
		*out3 = append(*out3, 3)
		return nil
	}, Once(), Async())

	results1 := []int{}
	results2 := []int{}
	results3 := []int{}
	e.Fire3(&results1, &results2, &results3)
	e.Fire3(&results1, &results2, &results3)
	e.Fire3(&results1, &results2, &results3)
	e.WaitAsync()

	assert.Len(t, results1, 1)
	assert.Len(t, results2, 1)
	assert.Len(t, results3, 1)
	assert.False(t, e.HasHandlers())
}

func TestEvent3OnAsync(t *testing.T) {
	e := E3[chan<- int, chan<- int, chan<- int]{}
	e.On(func(out1, out2, out3 chan<- int) error {
		out1 <- 1
		close(out1)
		out2 <- 2
		close(out2)
		out3 <- 3
		close(out3)
		return nil
	}, Async())

	results1 := make(chan int, 1)
	results2 := make(chan int, 1)
	results3 := make(chan int, 1)
	e.Fire3(results1, results2, results3)

	var numResults int64 = 0
	go func() {
		for range results1 {
			atomic.AddInt64(&numResults, 1)
		}
		for range results2 {
			atomic.AddInt64(&numResults, 1)
		}
		for range results3 {
			atomic.AddInt64(&numResults, 1)
		}
	}()
	e.WaitAsync()

	assert.Eventually(t, func() bool {
		return atomic.LoadInt64(&numResults) == 3
	}, 1*time.Second, 10*time.Millisecond)
}

func BenchmarkEvent3FireIntIntIntArg(b *testing.B) {
	e := E3[int, int, int]{}
	timesCalled := 0
	handler := func(_, _, _ int) error { timesCalled++; return nil }
	e.On(handler)
	for b.Loop() {
		e.Fire3(b.N, b.N, b.N)
	}

	assert.Equal(b, b.N, timesCalled)
}
