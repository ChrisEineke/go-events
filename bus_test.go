package EventBus

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	bus := New()
	if bus == nil {
		t.Log("New EventBus not created!")
		t.Fail()
	}
}

func TestHasCallback(t *testing.T) {
	bus := New()
	bus.Topic("topic1").On(func() {})
	bus.Topic("topic2")

	if !bus.Topic("topic1").Fireable() {
		t.Fail()
	}
	if bus.Topic("topic2").Fireable() {
		t.Fail()
	}
}

func TestOn(t *testing.T) {
	bus := New()
	if bus.Topic("topic").On(func() {}) != nil {
		t.Fail()
	}
	if bus.Topic("topic").On("String") == nil {
		t.Fail()
	}
}

func TestOnOnce(t *testing.T) {
	bus := New()
	if bus.Topic("topic").On(func() {}, Once()) != nil {
		t.Fail()
	}
	if bus.Topic("topic").On("String", Once()) == nil {
		t.Fail()
	}
}

func TestOnOnceAndManyOn(t *testing.T) {
	bus := New()
	event := "topic"
	flag := 0
	fn := func() { flag += 1 }
	bus.Topic(event).On(fn, Once())
	bus.Topic(event).On(fn)
	bus.Topic(event).On(fn)
	bus.Topic(event).Fire(event)

	if flag != 3 {
		t.Fail()
	}
}

func TestManyOnOnce(t *testing.T) {
	bus := New()
	event := "topic"
	var flags [3]byte

	bus.Topic(event).On(func() { flags[0]++ }, Once())
	bus.Topic(event).On(func() { flags[1]++ }, Once())
	bus.Topic(event).On(func() { flags[2]++ })

	bus.Topic(event).Fire()
	bus.Topic(event).Fire()

	if flags != [3]byte{1, 1, 2} {
		t.Fail()
	}
}

func TestOff(t *testing.T) {
	bus := New()
	handler := func() {}
	bus.Topic("topic").On(handler)
	if bus.Topic("topic").Off(handler) != nil {
		t.Fail()
	}
	if bus.Topic("topic").Off(handler) == nil {
		t.Fail()
	}
}

type handler struct {
	val int
}

func (h *handler) Handle() {
	h.val++
}

func TestOffMethod(t *testing.T) {
	bus := New()
	h := &handler{val: 0}

	bus.Topic("topic").On(h.Handle)
	bus.Topic("topic").Fire()
	if bus.Topic("topic").Off(h.Handle) != nil {
		t.Fail()
	}
	if bus.Topic("topic").Off(h.Handle) == nil {
		t.Fail()
	}
	bus.Topic("topic").Fire()
	bus.WaitAsync()

	if h.val != 1 {
		t.Fail()
	}
}

func TestFire(t *testing.T) {
	bus := New()
	bus.Topic("topic").On(func(a int, err error) {
		if a != 10 {
			t.Fail()
		}

		if err != nil {
			t.Fail()
		}
	})
	bus.Topic("topic").Fire(10, nil)
}

func TestSubscribeOnceAsync(t *testing.T) {
	results := make([]int, 0)

	bus := New()
	bus.Topic("topic").On(func(a int, out *[]int) {
		*out = append(*out, a)
	}, Once(), Async())

	bus.Topic("topic").Fire(10, &results)
	bus.Topic("topic").Fire(10, &results)
	bus.WaitAsync()

	if len(results) != 1 {
		t.Fail()
	}

	if bus.Topic("topic").Fireable() {
		t.Fail()
	}
}

func TestOnAsyncTransactional(t *testing.T) {
	results := make([]int, 0)

	bus := New()
	bus.Topic("topic").On(func(a int, out *[]int, dur string) {
		sleep, _ := time.ParseDuration(dur)
		time.Sleep(sleep)
		*out = append(*out, a)
	}, Async(), Transactional())

	bus.Topic("topic").Fire(1, &results, "1s")
	bus.Topic("topic").Fire(2, &results, "0s")
	bus.WaitAsync()

	if len(results) != 2 {
		t.Fail()
	}

	if results[0] != 1 || results[1] != 2 {
		t.Fail()
	}
}

func TestOnAsync(t *testing.T) {
	results := make(chan int)

	bus := New()
	bus.Topic("topic").On(func(a int, out chan<- int) {
		out <- a
	}, Async())

	bus.Topic("topic").Fire(1, results)
	bus.Topic("topic").Fire(2, results)

	numResults := 0

	go func() {
		for range results {
			numResults++
		}
	}()

	bus.WaitAsync()

	time.Sleep(10 * time.Millisecond)

	// todo race detected during execution of test
	//if numResults != 2 {
	//	t.Fail()
	//}
}

func TestCallbackArgsMismatch(t *testing.T) {
	bus := New()
	bus.Topic("topic").On(func(a int) {})
	bus.Topic("topic").Fire(1, 2)
}

func BenchmarkFireNoArgs(b *testing.B) {
	bus := New()
	callback := func() {}
	topic := bus.Topic("topic")
	topic.On(callback)
	for b.Loop() {
		topic.Fire()
	}
}

func BenchmarkFireIntArg(b *testing.B) {
	bus := New()
	callback := func(_ int) {}
	topic := bus.Topic("topic")
	topic.On(callback)
	for b.Loop() {
		topic.Fire(b.N)
	}
}
