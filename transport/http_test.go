package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charlinchui/galliard/message"
	"github.com/charlinchui/galliard/server"
)

func TestHTTPHandler_HandshakeSubscribePublish(t *testing.T) {
	srv := server.NewServer()
	handler := NewHTTPHandler(srv)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Handshake
	handshakeReq := []message.BayeuxMessage{{Channel: "/meta/handshake"}}
	handshakeResp := postBayeux(t, ts.URL, handshakeReq)
	if len(handshakeResp) != 1 || handshakeResp[0].ClientID == "" {
		t.Fatalf("handshake failed: %+v", handshakeResp)
	}
	clientID := handshakeResp[0].ClientID

	// Subscribe
	subscribeReq := []message.BayeuxMessage{{
		Channel:      "/meta/subscribe",
		ClientID:     clientID,
		Subscription: "/foo",
	}}
	subscribeResp := postBayeux(t, ts.URL, subscribeReq)
	if len(subscribeResp) != 1 || subscribeResp[0].Successful == nil || !*subscribeResp[0].Successful {
		t.Fatalf("subscribe failed: %+v", subscribeResp)
	}

	// Publish
	publishReq := []message.BayeuxMessage{{
		Channel:  "/foo",
		ClientID: clientID,
		Data:     map[string]interface{}{"msg": "hello"},
	}}
	publishResp := postBayeux(t, ts.URL, publishReq)
	if len(publishResp) != 1 || publishResp[0].Successful == nil || !*publishResp[0].Successful {
		t.Fatalf("publish failed: %+v", publishResp)
	}

	// Connect (should receive the published message)
	connectReq := []message.BayeuxMessage{{
		Channel:  "/meta/connect",
		ClientID: clientID,
	}}
	connectResp := postBayeux(t, ts.URL, connectReq)
	if len(connectResp) != 1 || connectResp[0].Channel != "/foo" {
		t.Fatalf("connect did not deliver published message: %+v", connectResp)
	}
	if connectResp[0].Data["msg"] != "hello" {
		t.Errorf("expected 'hello', got %v", connectResp[0].Data["msg"])
	}
}

func postBayeux(t *testing.T, url string, msgs []message.BayeuxMessage) []message.BayeuxMessage {
	data, err := json.Marshal(msgs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out []message.BayeuxMessage
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal: %v\nbody: %s", err, string(body))
	}
	return out
}
