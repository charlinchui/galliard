package message

import (
	"encoding/json"
	"testing"
)

func TestMarshalUnmarshalBayeuxMessage(t *testing.T) {
	original := BayeuxMessage{
		Channel:      "/meta/handshake",
		ClientID:     "abc123",
		Data:         map[string]interface{}{"foo": "bar"},
		ID:           "msg1",
		Subscription: "/chat/room1",
		Successful:   boolPtr(true),
		Error:        "",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded BayeuxMessage

	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Channel != original.Channel ||
		decoded.ClientID != original.ClientID ||
		decoded.ID != original.ID ||
		decoded.Subscription != original.Subscription ||
		decoded.Successful == nil ||
		*decoded.Successful != *original.Successful {
		t.Errorf("Decoded message does not match original message")
	}

	if decoded.Data["foo"] != "bar" {
		t.Errorf("Decoded data does not match original message")
	}
}

func boolPtr(b bool) *bool { return &b }
