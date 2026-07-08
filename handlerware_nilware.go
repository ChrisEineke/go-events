package events

type Nilware struct{ Handlerware }

func (_ *Nilware) OnUse(Event) error                      { return nil }
func (_ *Nilware) OnDisuse(Event) error                   { return nil }
func (_ *Nilware) OnSubscribe(Event, Handler) error       { return nil }
func (_ *Nilware) OnUnsubscribe(Event, Handler) error     { return nil }
func (_ *Nilware) OnAllPreFire(Event, []any) error        { return nil }
func (_ *Nilware) OnPreFire(Event, Handler, []any) error  { return nil }
func (_ *Nilware) OnPostFire(Event, Handler, []any) error { return nil }
func (_ *Nilware) OnAllPostFire(Event, []any) error       { return nil }

var _ Handlerware = (*Nilware)(nil)
