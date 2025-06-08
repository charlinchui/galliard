package transport

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/charlinchui/galliard/message"
	"github.com/charlinchui/galliard/server"
)

type HTTPHandler struct {
	Server *server.Server
}

func NewHTTPHandler(s *server.Server) *HTTPHandler {
	return &HTTPHandler{Server: s}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var reqMsgs []message.BayeuxMessage
	if err := json.Unmarshal(body, &reqMsgs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var respMsgs []message.BayeuxMessage
	for i := range reqMsgs {
		resp := h.Server.HandleMessage(&reqMsgs[i])
		if resp != nil {
			respMsgs = append(respMsgs, *resp)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respMsgs)
}
