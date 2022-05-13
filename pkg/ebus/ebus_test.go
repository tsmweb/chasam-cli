package ebus

import (
	"testing"
	"time"
)

func TestEventBus(t *testing.T) {
	u1 := user{id: 1, name: "user-1"}
	u2 := user{id: 2, name: "user-2"}
	u3 := user{id: 3, name: "user-3"}

	ready := make(chan struct{})

	go func() {
		sub := Instance().Subscribe(topic)
		defer sub.Unsubscribe()

		close(ready)

		for e := range sub.Event {
			//go func(e DataEvent) {
			u := e.Data.(user)
			t.Logf("Topic: %s", e.Topic)
			t.Logf("Data: %v", u)

			if u.id != u1.id && u.id != u2.id && u.id != u3.id {
				t.Errorf("user.id = %d, want = %d or %d or %d", u.id, u1.id, u2.id, u3.id)
			}
			//}(event)
		}
	}()

	<-ready

	Instance().Publish(topic, u1)
	Instance().Publish(topic, u2)
	Instance().Publish(topic, u3)

	time.Sleep(time.Millisecond * 100)
}

const topic = "user"

type user struct {
	id   int
	name string
}
