package events

type Nilware struct{ Handlerware }

func (_ *Nilware) OnUse(*E) error                      { return nil }
func (_ *Nilware) OnDisuse(*E) error                   { return nil }
func (_ *Nilware) OnSubscribe(*E, Handler) error       { return nil }
func (_ *Nilware) OnUnsubscribe(*E, Handler) error     { return nil }
func (_ *Nilware) OnAllPreFire(*E, []any) error        { return nil }
func (_ *Nilware) OnPreFire(*E, Handler, []any) error  { return nil }
func (_ *Nilware) OnPostFire(*E, Handler, []any) error { return nil }
func (_ *Nilware) OnAllPostFire(*E, []any) error       { return nil }
