package events

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvent0HasHandlers(t *testing.T) {
	e := E0{}
	assert.Equal(t, e.HasHandlers(), false, "there should be no Handlers")

	e.On(func() error { return nil })
	assert.Equal(t, e.HasHandlers(), true, "there should be a Handlers")
}

func TestEvent0On(t *testing.T) {
	e := E0{}

	err := e.On(func() error { return nil })
	assert.NoError(t, err)
}

func TestEvent0Off(t *testing.T) {
	e := E0{}
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

func TestEvent0Fire(t *testing.T) {
	e := E0{}
	e.On(func() error {
		return nil
	})
	e.Fire(1)
}

func TestEvent0Fire0(t *testing.T) {
	e := E0{}
	e.On(func() error {
		return nil
	})
	e.Fire0()
}

func TestEvent0Fire0WithHandlerware(t *testing.T) {
	e := E0{}
	tw := &testware{}
	callable := func() error { return nil }

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire0()
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable)
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire0()
	assert.Equal(t, 2, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 1, tw.onPreFireCalled, "OnPreFire should be called for every Handler")
	assert.Equal(t, 1, tw.onPostFireCalled, "OnPostFire should be called for every Handler")
	assert.Equal(t, 2, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.Off(callable)
	assert.Equal(t, 1, tw.onUnsubscribeCalled, "OnUnsubscribe should be called once for every callable detached from the Event")

	e.Disuse(tw)
	assert.Equal(t, 1, tw.onDisuseCalled)
}

func TestEvent0Fire0AsyncWithHandlerware(t *testing.T) {
	e := E0{}
	tw := &testware{}
	callable := func() error { return nil }

	e.Use(tw)
	assert.Equal(t, 1, tw.onUseCalled)

	e.Fire0()
	assert.Equal(t, 1, tw.onAllPreFireCalled, "OnAllPreFire should be called once even if there are no Handlers")
	assert.Equal(t, 0, tw.onPreFireCalled, "OnPreFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 0, tw.onPostFireCalled, "OnPostFire shouldn't be called since there are no Handlers")
	assert.Equal(t, 1, tw.onAllPostFireCalled, "OnAllPostFire should be called once even if there are no Handlers")

	e.On(callable, Async())
	assert.Equal(t, 1, tw.onSubscribeCalled, "OnSubscribe should be called once for every callable attached to the Event")

	e.Fire0()
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

func TestEvent0OnOnceAndManyOn(t *testing.T) {
	e := E0{}
	flag := 0
	fn := func() error { flag += 1; return nil }
	e.On(fn, Once())
	e.On(fn)
	e.On(fn)
	e.Fire0()

	assert.Equal(t, flag, 3)
}

func TestEvent0ManyOnOnce(t *testing.T) {
	e := E0{}
	var flags [3]byte

	e.On(func() error { flags[0]++; return nil }, Once())
	e.On(func() error { flags[1]++; return nil }, Once())
	e.On(func() error { flags[2]++; return nil })

	e.Fire0()
	e.Fire0()

	assert.Equal(t, flags, [3]byte{1, 1, 2})
}

func TestEvent0OnOffFunction(t *testing.T) {
	e := E0{}
	handler := func() error { return nil }

	e.On(handler)
	err := e.Off(handler)
	assert.NoError(t, err)

	err = e.Off(handler)
	assert.Error(t, err)
}

type testHandler0 struct {
	val int
}

func (h *testHandler0) Handle() error {
	h.val++
	return nil
}

func TestEvent0OnOffReceiver(t *testing.T) {
	e := E0{}
	handler := &testHandler0{val: 0}

	e.On(handler.Handle)
	e.Fire0()
	err := e.Off(handler.Handle)
	assert.NoError(t, err)

	err = e.Off(handler.Handle)
	assert.Error(t, err)

	e.Fire0()
	e.WaitAsync()
	assert.Equal(t, 1, handler.val, "handler wasn't removed after calling Off")
}

func TestEvent0OnOnceAsync(t *testing.T) {
	results := []int{}

	e := E0{}
	e.On(func() error {
		results = append(results, 1)
		return nil
	}, Once(), Async())

	e.Fire0()
	e.Fire0()
	e.WaitAsync()

	assert.Len(t, results, 1)
	assert.False(t, e.HasHandlers())
}

func TestEvent0OnAsync(t *testing.T) {
	results := make(chan int, 2)

	e := E0{}
	e.On(func() error {
		results <- 1
		close(results)
		return nil
	}, Async())

	e.Fire0()

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

func BenchmarkEvent0FireIntArg(b *testing.B) {
	e := E0{}
	timesCalled := 0
	handler := func() error { timesCalled++; return nil }
	e.On(handler)
	for b.Loop() {
		e.Fire0()
	}

	assert.Equal(b, b.N, timesCalled)
}
