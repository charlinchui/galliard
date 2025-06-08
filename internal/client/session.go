package client

import (
	"sync"

	"github.com/charlinchui/galliard/message"
)

type Session struct {
	ID            string
	Subscriptions map[string]struct{}
	MessageQueue  []*message.BayeuxMessage
	Advice        *message.Advice
	mu            sync.Mutex
}

func (s *Session) SetAdvice(advice *message.Advice) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Advice = advice
}

func (s *Session) GetAdvice() *message.Advice {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Advice
}

func NewSession(id string) *Session {
	return &Session{
		ID:            id,
		Subscriptions: make(map[string]struct{}),
		MessageQueue:  []*message.BayeuxMessage{},
	}
}

func (s *Session) Subscribe(channel string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Subscriptions[channel] = struct{}{}
}

func (s *Session) Unsubscribe(channel string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Subscriptions, channel)
}

func (s *Session) IsSubscribed(channel string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.Subscriptions[channel]
	return ok
}

func (s *Session) Enqueue(msg *message.BayeuxMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.MessageQueue = append(s.MessageQueue, msg)
}

func (s *Session) DequeueAll() []*message.BayeuxMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	msgs := s.MessageQueue
	s.MessageQueue = []*message.BayeuxMessage{}
	return msgs
}
