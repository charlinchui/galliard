package channel

import (
	"sync"

	"github.com/charlinchui/galliard/client"
	"github.com/charlinchui/galliard/message"
)

type Channel struct {
	Name        string
	Subscribers map[string]*client.Session
	mu          sync.Mutex
}

func NewChannel(name string) *Channel {
	return &Channel{
		Name:        name,
		Subscribers: make(map[string]*client.Session),
	}
}

func (ch *Channel) Subscribe(s *client.Session) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.Subscribers[s.ID] = s
}

func (ch *Channel) Unsubscribe(s *client.Session) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	delete(ch.Subscribers, s.ID)
}

func (ch *Channel) Publish(msg *message.BayeuxMessage) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	for _, s := range ch.Subscribers {
		s.Enqueue(msg)
	}
}
