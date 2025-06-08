package server

import (
	"testing"

	"github.com/charlinchui/galliard/message"
)

func TestNewServer(t *testing.T) {
	srv := NewServer()
	if srv == nil {
		t.Fatal("Expected non-nil server")
	}
	if len(srv.Sessions) != 0 {
		t.Errorf("Expected no sessions")
	}
	if len(srv.Channels) != 0 {
		t.Errorf("Expected no channels")
	}
}

func TestRegisterNewSession(t *testing.T) {
	srv := NewServer()
	id := "client-xyz"
	s := srv.RegisterSession(id)
	if s.ID != id {
		t.Errorf("Expected session ID to be client-xyz")
	}
	if got := srv.GetSession(id); got != s {
		t.Errorf("Expected GetSession to return the registered session")
	}
}

func TestGetOrCreateChannel(t *testing.T) {
	srv := NewServer()
	ch1 := srv.GetOrCreateChannel("/foo")
	if ch1.Name != "/foo" {
		t.Errorf("Expected channel name to be '/foo'")
	}
	ch2 := srv.GetOrCreateChannel("/foo")
	if ch1 != ch2 {
		t.Errorf("Expected the same channel to be returned")
	}
}

func TestHandleHandshake(t *testing.T) {
	srv := NewServer()
	req := &message.BayeuxMessage{Channel: "/meta/handshake"}
	resp := srv.HandleMessage(req)
	if resp.Channel != "/meta/handshake" {
		t.Errorf("Expected channel name to be /meta/handshake")
	}
	if resp.ClientID == "" {
		t.Errorf("Expected non-empty clientID")
	}
	if resp.Successful == nil || !*resp.Successful {
		t.Errorf("Expected successful handshake")
	}
	if got := srv.GetSession(resp.ClientID); got == nil {
		t.Errorf("Expected session to be registered after handshake")
	}
}

func TestHandleAndSubscribe(t *testing.T) {
	srv := NewServer()
	handshake := &message.BayeuxMessage{Channel: "/meta/handshake"}
	resp := srv.HandleMessage(handshake)
	clientID := resp.ClientID

	subscribe := &message.BayeuxMessage{
		Channel:      "/meta/subscribe",
		ClientID:     clientID,
		Subscription: "/foo",
	}
	subResp := srv.HandleMessage(subscribe)
	if subResp.Successful == nil || !*subResp.Successful {
		t.Errorf("Expected successful subscription")
	}

	publish := &message.BayeuxMessage{
		Channel:  "/foo",
		ClientID: clientID,
		Data: map[string]interface{}{
			"msg": "hello",
		},
	}
	pubResp := srv.HandleMessage(publish)
	if pubResp.Successful == nil || !*pubResp.Successful {
		t.Errorf("Expected successful publish")
	}

	connect := &message.BayeuxMessage{
		Channel:  "/meta/connect",
		ClientID: clientID,
	}
	connResp := srv.HandleMessage(connect)
	if connResp.Channel != "/foo" {
		t.Errorf("Expected message from '/foo', got %q", connResp.Channel)
	}
	if connResp.Data["msg"] != "hello" {
		t.Errorf("Expected data 'hello', got %v", connResp.Data["msg"])
	}
}

func TestHandleUnsubscribe(t *testing.T) {
	srv := NewServer()
	handshake := &message.BayeuxMessage{Channel: "/meta/handshake"}
	resp := srv.HandleMessage(handshake)
	clientID := resp.ClientID

	srv.HandleMessage(&message.BayeuxMessage{
		Channel:      "/meta/subscribe",
		ClientID:     clientID,
		Subscription: "/bar",
	})
	unsubResp := srv.HandleMessage(&message.BayeuxMessage{
		Channel:      "/meta/unsubscribe",
		ClientID:     clientID,
		Subscription: "/bar",
	})
	if unsubResp.Successful == nil || !*unsubResp.Successful {
		t.Errorf("Expected subscription to end successfully")
	}
}

func TestHandleDisconnect(t *testing.T) {
	srv := NewServer()
	handshake := &message.BayeuxMessage{Channel: "/meta/handshake"}
	resp := srv.HandleMessage(handshake)
	clientID := resp.ClientID

	dcResp := srv.HandleMessage(&message.BayeuxMessage{
		Channel:  "/meta/disconnect",
		ClientID: clientID,
	})

	if dcResp.Successful == nil || !*dcResp.Successful {
		t.Errorf("Expected successful disconnect")
	}

	if got := srv.GetSession(clientID); got != nil {
		t.Errorf("Expected session to be removed after disconnect")
	}
}

func TestMessageIDCorrelation(t *testing.T) {
	srv := NewServer()
	handshake := &message.BayeuxMessage{Channel: "/meta/handshake", ID: "h1"}
	resp := srv.HandleMessage(handshake)
	if resp.ID != "h1" {
		t.Errorf("expected id 'h1', got %q", resp.ID)
	}

	clientID := resp.ClientID
	subscribe := &message.BayeuxMessage{
		Channel:      "/meta/subscribe",
		ClientID:     clientID,
		Subscription: "/foo",
		ID:           "s1",
	}
	subResp := srv.HandleMessage(subscribe)
	if subResp.ID != "s1" {
		t.Errorf("expected id 's1', got %q", subResp.ID)
	}
}

func TestErrorHandling_MissingFields(t *testing.T) {
	srv := NewServer()
	connect := &message.BayeuxMessage{Channel: "/meta/connect", ID: "c1"}
	resp := srv.HandleMessage(connect)
	if resp.Successful == nil || *resp.Successful != false {
		t.Errorf("expected unsuccessful response")
	}
	if resp.Error == "" {
		t.Errorf("expected error message")
	}
	if resp.ID != "c1" {
		t.Errorf("expected id 'c1', got %q", resp.ID)
	}

	handshake := &message.BayeuxMessage{Channel: "/meta/handshake"}
	hresp := srv.HandleMessage(handshake)
	clientID := hresp.ClientID
	subscribe := &message.BayeuxMessage{
		Channel:  "/meta/subscribe",
		ClientID: clientID,
		ID:       "s1",
	}
	sresp := srv.HandleMessage(subscribe)
	if sresp.Successful == nil || *sresp.Successful != false {
		t.Errorf("expected unsuccessful response")
	}
	if sresp.Error == "" {
		t.Errorf("expected error message")
	}
	if sresp.ID != "s1" {
		t.Errorf("expected id 's1', got %q", sresp.ID)
	}
}

func TestErrorHandling_UnknownClient(t *testing.T) {
	srv := NewServer()
	subscribe := &message.BayeuxMessage{
		Channel:      "/meta/subscribe",
		ClientID:     "not-a-client",
		Subscription: "/foo",
		ID:           "s2",
	}
	resp := srv.HandleMessage(subscribe)
	if resp.Successful == nil || *resp.Successful != false {
		t.Errorf("expected unsuccessful response")
	}
	if resp.Error == "" {
		t.Errorf("expected error message")
	}
	if resp.ID != "s2" {
		t.Errorf("expected id 's2', got %q", resp.ID)
	}

}
