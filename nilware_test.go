package events

import "testing"

func TestNilwareImplementsInterfaces(t *testing.T) {
	var _ Handlerware = &Nilware{}
}

func TestNilware(t *testing.T) {
	nw := &Nilware{}
	nw.OnUse(nil)
	nw.OnDisuse(nil)
	nw.OnSubscribe(nil, nil)
	nw.OnUnsubscribe(nil, nil)
	nw.OnAllPreFire(nil, nil)
	nw.OnPreFire(nil, nil, nil)
	nw.OnPostFire(nil, nil, nil)
	nw.OnAllPostFire(nil, nil)
}
