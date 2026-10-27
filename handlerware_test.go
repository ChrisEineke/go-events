package events

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

type testHandler struct {
	val int
}

func (h *testHandler) Handle() error {
	h.val++
	return nil
}
