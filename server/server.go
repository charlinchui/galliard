package server

import (
	"sync"

	"github.com/charlinchui/galliard/channel"
	"github.com/charlinchui/galliard/client"
	"github.com/charlinchui/galliard/message"
	"github.com/charlinchui/galliard/utils"
)

type Server struct {
	Sessions map[string]*client.Session
	Channels map[string]*channel.Channel
	mu       sync.Mutex
}

func NewServer() *Server {
	return &Server{
		Sessions: make(map[string]*client.Session),
		Channels: make(map[string]*channel.Channel),
	}
}

func (s *Server) RegisterSession(id string) *client.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	session := client.NewSession(id)
	s.Sessions[id] = session
	return session
}

func (s *Server) GetSession(id string) *client.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Sessions[id]
}

func (s *Server) GetOrCreateChannel(chName string) *channel.Channel {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.Channels[chName]
	if !ok {
		ch = channel.NewChannel(chName)
		s.Channels[chName] = ch
	}
	return ch
}

func (s *Server) HandleMessage(msg *message.BayeuxMessage) *message.BayeuxMessage {
	if errResp := validateMessage(msg, s); errResp != nil {
		return errResp
	}
	switch msg.Channel {
	case "/meta/handshake":
		return s.handleHandshake(msg)
	case "/meta/connect":
		return s.handleConnect(msg)
	case "/meta/subscribe":
		return s.handleSubscribe(msg)
	case "/meta/unsubscribe":
		return s.handleUnsubscribe(msg)
	case "/meta/disconnect":
		return s.handleDisconnect(msg)
	default:
		return s.handlePublish(msg)
	}
}

func (s *Server) handleHandshake(msg *message.BayeuxMessage) *message.BayeuxMessage {
	clientID := utils.GenerateID()
	s.RegisterSession(clientID)
	success := true
	return &message.BayeuxMessage{
		Channel:    "/meta/handshake",
		Successful: &success,
		ClientID:   clientID,
		ID:         msg.ID,
	}
}

func (s *Server) handleConnect(msg *message.BayeuxMessage) *message.BayeuxMessage {
	sess := s.GetSession(msg.ClientID)
	queued := sess.DequeueAll()
	if len(queued) > 0 {
		return queued[0]
	}
	return &message.BayeuxMessage{
		Channel:    "/meta/connect",
		ClientID:   sess.ID,
		Successful: nil,
		ID:         msg.ID,
	}
}

func (s *Server) handleSubscribe(msg *message.BayeuxMessage) *message.BayeuxMessage {
	sess := s.GetSession(msg.ClientID)
	ch := s.GetOrCreateChannel(msg.Subscription)
	ch.Subscribe(sess)
	sess.Subscribe(msg.Subscription)
	success := true
	return &message.BayeuxMessage{
		Channel:      "/meta/subscribe",
		Successful:   &success,
		Subscription: msg.Subscription,
		ID:           msg.ID,
	}
}

func (s *Server) handleUnsubscribe(msg *message.BayeuxMessage) *message.BayeuxMessage {
	ch := s.GetOrCreateChannel(msg.Subscription)
	sess := s.GetSession(msg.ClientID)
	ch.Unsubscribe(sess)
	sess.Unsubscribe(msg.Subscription)
	success := true
	return &message.BayeuxMessage{
		Channel:      "/meta/unsubscribe",
		Successful:   &success,
		Subscription: msg.Subscription,
		ID:           msg.ID,
	}
}

func (s *Server) handleDisconnect(msg *message.BayeuxMessage) *message.BayeuxMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.Sessions[msg.ClientID]
	if ok {
		for sub := range sess.Subscriptions {
			if ch, exists := s.Channels[sub]; exists {
				ch.Unsubscribe(sess)
			}
		}
		delete(s.Sessions, msg.ClientID)
	}
	success := true
	return &message.BayeuxMessage{
		Channel:    "/meta/disconnect",
		Successful: &success,
		ID:         msg.ID,
	}
}

func (s *Server) handlePublish(msg *message.BayeuxMessage) *message.BayeuxMessage {
	ch := s.GetOrCreateChannel(msg.Channel)
	ch.Publish(msg)
	success := true
	return &message.BayeuxMessage{
		Channel:    msg.Channel,
		Successful: &success,
		ID:         msg.ID,
	}
}

func errorResponse(channel, id, errMsg string) *message.BayeuxMessage {
	success := false
	return &message.BayeuxMessage{
		Channel:    channel,
		Successful: &success,
		Error:      errMsg,
		ID:         id,
	}
}

func validateMessage(msg *message.BayeuxMessage, s *Server) *message.BayeuxMessage {
	switch msg.Channel {
	case "/meta/handshake":
		return nil
	case "/meta/connect", "/meta/disconnect":
		if msg.ClientID == "" {
			return errorResponse(msg.Channel, msg.ID, "Missing clientId")
		}
		if s.GetSession(msg.ClientID) == nil {
			return errorResponse(msg.Channel, msg.ID, "Unknown ClientID")
		}
	case "/meta/subscribe", "/meta/unsubscribe":
		if msg.ClientID == "" {
			return errorResponse(msg.Channel, msg.ID, "Missing clientId")
		}
		if msg.Subscription == "" {
			return errorResponse(msg.Channel, msg.ID, "Missing subscription")
		}
		if s.GetSession(msg.ClientID) == nil {
			return errorResponse(msg.Channel, msg.ID, "Unknown ClientID")
		}
	default:
		if msg.ClientID == "" {
			return errorResponse(msg.Channel, msg.ID, "Missing clientId")
		}
		if msg.Channel == "" {
			return errorResponse(msg.Channel, msg.ID, "Missing channel")
		}
		if s.GetSession(msg.ClientID) == nil {
			return errorResponse(msg.Channel, msg.ID, "Unknown ClientID")
		}
	}
	return nil
}
