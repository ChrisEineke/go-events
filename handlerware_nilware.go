package events

type Nilware struct{ Handlerware }

func (_ *Nilware) OnUse(EventSource) error                       { return nil }
func (_ *Nilware) OnDisuse(EventSource) error                    { return nil }
func (_ *Nilware) OnSubscribe(EventSource, Handler) error        { return nil }
func (_ *Nilware) OnUnsubscribe(EventSource, Handler) error      { return nil }
func (_ *Nilware) OnAllPreFire(EventSource, ...any) error        { return nil }
func (_ *Nilware) OnPreFire(EventSource, Handler, ...any) error  { return nil }
func (_ *Nilware) OnPostFire(EventSource, Handler, ...any) error { return nil }
func (_ *Nilware) OnAllPostFire(EventSource, ...any) error       { return nil }

var _ Handlerware = (*Nilware)(nil)
