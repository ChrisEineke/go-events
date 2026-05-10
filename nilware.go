package events

type Nilware struct{ Handlerware }

func (_ *Nilware) OnUse(*E) error                { return nil }
func (_ *Nilware) OnDisuse(*E) error             { return nil }
func (_ *Nilware) OnSubscribe(*E, Handler)       {}
func (_ *Nilware) OnUnsubscribe(*E, Handler)     {}
func (_ *Nilware) OnAllPreFire(*E, []any)        {}
func (_ *Nilware) OnPreFire(*E, Handler, []any)  {}
func (_ *Nilware) OnPostFire(*E, Handler, []any) {}
func (_ *Nilware) OnAllPostFire(*E, []any)       {}
