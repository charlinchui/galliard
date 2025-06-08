package message

// Advice provides connection and reconnection instructions to Bayeux clients.
// It is typically included in /meta/handshake and /meta/connect responses
// to inform clients how to handle reconnection, intervals, and timeouts.
type Advice struct {
	// Reconnect specifies the reconnection strategy for the client.
	// Common values are "retry", "handshake", or "none".
	Reconnect string `json:"reconnect,omitempty"`

	// Interval is the number of milliseconds the client should wait before reconnecting.
	Interval int `json:"interval,omitempty"`

	// Timeout is the maximum time in milliseconds the server will hold a long-polling request.
	Timeout int `json:"timeout,omitempty"`
}

// BayeuxMessage represents a message in the Bayeux protocol.
// It is used for all communication between clients and the server,
// including meta operations (handshake, connect, subscribe, etc.)
// and data messages on application channels.
type BayeuxMessage struct {
	// Channel is the destination or meta channel for the message (e.g., "/meta/handshake", "/foo/bar").
	Channel string `json:"channel"`

	// ClientID is the unique identifier for the client session, assigned by the server during handshake.
	ClientID string `json:"clientId,omitempty"`

	// Data contains the payload for publish messages or additional information for meta messages.
	Data map[string]interface{} `json:"data,omitempty"`

	// ID is an optional unique identifier for correlating requests and responses.
	ID string `json:"id,omitempty"`

	// Subscription specifies the channel to subscribe or unsubscribe to (used in /meta/subscribe and /meta/unsubscribe).
	Subscription string `json:"subscription,omitempty"`

	// Successful indicates whether a meta operation was successful.
	Successful *bool `json:"successful,omitempty"`

	// Error contains an error message if the operation failed.
	Error string `json:"error,omitempty"`

	// Advice provides connection advice to the client, typically included in handshake and connect responses.
	Advice *Advice `json:"advice,omitempty"`
}
