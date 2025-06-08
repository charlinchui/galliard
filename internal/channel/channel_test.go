package channel

import (
	"testing"

	"github.com/charlinchui/galliard/internal/client"
	"github.com/charlinchui/galliard/message"
)

func newTestSession(id string) *client.Session {
	return client.NewSession(id)
}

func TestNewChannel(t *testing.T) {
	ch := NewChannel("/foo")
	if ch.Name != "/foo" {
		t.Errorf("Expected name /foo, got %q", ch.Name)
	}
	if len(ch.Subscribers) != 0 {
		t.Errorf("Expected no subscribers")
	}
}

func TestSubscribeUnsubscribe(t *testing.T) {
	ch := NewChannel("/bar")
	s1 := newTestSession("c1")
	s2 := newTestSession("c2")

	ch.Subscribe(s1)
	ch.Subscribe(s2)
	if len(ch.Subscribers) != 2 {
		t.Errorf("Expected 2 subscribers")
	}
	ch.Unsubscribe(s1)
	if len(ch.Subscribers) != 1 {
		t.Errorf("Expected 1 subscriber afet unsubscribe")
	}

	if _, ok := ch.Subscribers[s2.ID]; !ok {
		t.Errorf("s2 should still be subscribed")
	}
}

func TestPublish(t *testing.T) {
	ch := NewChannel("/bar")
	s1 := newTestSession("c1")
	s2 := newTestSession("c2")

	ch.Subscribe(s1)
	ch.Subscribe(s2)
	msg := &message.BayeuxMessage{
		Channel: "/baz",
		Data: map[string]interface{}{
			"x": 1.0,
		},
	}
	ch.Publish(msg)
	msgs1 := s1.DequeueAll()
	msgs2 := s2.DequeueAll()

	if len(msgs1) != 1 || len(msgs2) != 1 {
		t.Errorf("Expected 1 message per subscriber")
	}
	if msgs1[0].Channel != "/baz" || msgs2[0].Channel != "/baz" {
		t.Errorf("Expected channel to be '/baz'")
	}
	if msgs1[0].Data["x"] != 1.0 || msgs2[0].Data["x"] != 1.0 {
		t.Errorf("Expected data for 'x' to be 1.0")
	}
}
