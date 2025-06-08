// Package server provides a Bayeux protocol server implementation.
// It manages client sessions, channels, and message routing for real-time pub/sub systems.
package server

import (
	"sync"

	"github.com/charlinchui/galliard/internal/channel"
	"github.com/charlinchui/galliard/internal/client"
	"github.com/charlinchui/galliard/internal/utils"
	"github.com/charlinchui/galliard/message"
)

// Server implements a Bayeux protocol server.
// It manages client sessions, channels, and routes Bayeux messages.
type Server struct {
	Sessions   map[string]*client.Session
	Channels   map[string]*channel.Channel
	sessionsMu sync.RWMutex
	channelsMu sync.RWMutex
}

func defaultAdvice() *message.Advice {
	return &message.Advice{
		Reconnect: "retry",
		Interval:  0,
		Timeout:   10000,
	}
}

func (s *Server) setUpAdvice(msg *message.BayeuxMessage) *message.Advice {
	if msg != nil && msg.Advice != nil {
		return msg.Advice
	}
	sess := s.getSession(msg.ClientID)
	if sess != nil && sess.Advice != nil {
		return sess.Advice
	}
	return defaultAdvice()
}

// NewServer creates and returns a new Bayeux Server instance.
func NewServer() *Server {
	return &Server{
		Sessions: make(map[string]*client.Session),
		Channels: make(map[string]*channel.Channel),
	}
}

func (s *Server) registerSession(id string) *client.Session {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	sess := client.NewSession(id)
	s.Sessions[id] = sess
	return sess
}

func (s *Server) getSession(id string) *client.Session {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()
	return s.Sessions[id]
}

func (s *Server) getOrCreateChannel(chName string) *channel.Channel {
	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()
	ch, ok := s.Channels[chName]
	if !ok {
		ch = channel.NewChannel(chName)
		s.Channels[chName] = ch
	}
	return ch
}

// HandleMessage processes a BayeuxMessage and returns a response message.
// It handles all Bayeux meta channels and data publish messages.
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
	sess := s.registerSession(clientID)
	if msg.Advice != nil {
		sess.Advice = msg.Advice
	}
	success := true
	return &message.BayeuxMessage{
		Channel:    "/meta/handshake",
		Successful: &success,
		ClientID:   clientID,
		ID:         msg.ID,
		Advice:     s.setUpAdvice(msg),
	}
}

func (s *Server) handleConnect(msg *message.BayeuxMessage) *message.BayeuxMessage {
	sess := s.getSession(msg.ClientID)
	queued := sess.DequeueAll()
	if len(queued) > 0 {
		return queued[0]
	}
	return &message.BayeuxMessage{
		Channel:    "/meta/connect",
		ClientID:   sess.ID,
		Successful: nil,
		ID:         msg.ID,
		Advice:     s.setUpAdvice(msg),
	}
}

func (s *Server) handleSubscribe(msg *message.BayeuxMessage) *message.BayeuxMessage {
	sess := s.getSession(msg.ClientID)
	ch := s.getOrCreateChannel(msg.Subscription)
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
	ch := s.getOrCreateChannel(msg.Subscription)
	sess := s.getSession(msg.ClientID)
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
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	sess, ok := s.Sessions[msg.ClientID]
	if ok {
		s.channelsMu.Lock()
		defer s.channelsMu.Unlock()
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
	ch := s.getOrCreateChannel(msg.Channel)
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
		Advice:     defaultAdvice(),
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
		if s.getSession(msg.ClientID) == nil {
			return errorResponse(msg.Channel, msg.ID, "Unknown ClientID")
		}
	case "/meta/subscribe", "/meta/unsubscribe":
		if msg.ClientID == "" {
			return errorResponse(msg.Channel, msg.ID, "Missing clientId")
		}
		if msg.Subscription == "" {
			return errorResponse(msg.Channel, msg.ID, "Missing subscription")
		}
		if s.getSession(msg.ClientID) == nil {
			return errorResponse(msg.Channel, msg.ID, "Unknown ClientID")
		}
	default:
		if msg.ClientID == "" {
			return errorResponse(msg.Channel, msg.ID, "Missing clientId")
		}
		if msg.Channel == "" {
			return errorResponse(msg.Channel, msg.ID, "Missing channel")
		}
		if s.getSession(msg.ClientID) == nil {
			return errorResponse(msg.Channel, msg.ID, "Unknown ClientID")
		}
	}
	return nil
}
