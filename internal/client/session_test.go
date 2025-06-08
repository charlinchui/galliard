package client

import (
	"testing"

	"github.com/charlinchui/galliard/message"
)

func TestNewSession(t *testing.T) {
	s := NewSession("client-1")
	if s.ID != "client-1" {
		t.Errorf("Expected ID 'client-1, got %q", s.ID)
	}
	if len(s.Subscriptions) != 0 {
		t.Errorf("Expected no subscriptions")
	}
	if len(s.MessageQueue) != 0 {
		t.Errorf("Expected empty message queue")
	}
}

func TestSubscribeUnsubscribe(t *testing.T) {
	s := NewSession("client-2")
	s.Subscribe("/foo")
	if !s.IsSubscribed("/foo") {
		t.Errorf("Expected a subscription to /foo")
	}
	s.Unsubscribe("/foo")
	if s.IsSubscribed("/foo") {
		t.Errorf("Expected no subscription to /foo")
	}
}

func TestEnqueueDequeue(t *testing.T) {
	s := NewSession("client-2")
	msg := &message.BayeuxMessage{Channel: "/bar"}
	s.Enqueue(msg)
	if len(s.MessageQueue) != 1 {
		t.Errorf("Expected 1 message in queue")
	}

	msgs := s.DequeueAll()

	if len(msgs) != 1 || msgs[0].Channel != "/bar" {
		t.Errorf("Dequeued message mismatch")
	}
	if len(s.MessageQueue) != 0 {
		t.Errorf("Expected queue to be empty after dequeue")
	}
}

func TestSessionAdvice(t *testing.T) {
	s := NewSession("client-3")
	advice := &message.Advice{
		Reconnect: "retry",
		Interval:  1000,
		Timeout:   5000,
	}
	s.SetAdvice(advice)
	got := s.GetAdvice()
	if got == nil {
		t.Fatalf("Expected advice to be set")
	}
	if got.Reconnect != "retry" || got.Interval != 1000 || got.Timeout != 5000 {
		t.Errorf("Advice fields not set correctly: %+v", got)
	}
}
