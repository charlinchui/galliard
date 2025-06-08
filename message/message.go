package message

type BayeuxMessage struct {
	Channel      string                 `json:"channel"`
	ClientID     string                 `json:"clientId,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
	ID           string                 `json:"id,omitempty"`
	Subscription string                 `json:"subscription,omitempty"`
	Successful   *bool                  `json:"successful,omitempty"`
	Error        string                 `json:"error,omitempty"`
}
