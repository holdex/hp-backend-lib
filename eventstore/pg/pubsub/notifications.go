package pubsub

import (
	"sync"
)

type Notifications interface {
	Publish(Notification)
	Subscribe() (<-chan Notification, func())
}

func NewNotifications() Notifications {
	return &notifications{subscribers: map[chan Notification]bool{}}
}

type notifications struct {
	sync.RWMutex
	subscribers map[chan Notification]bool
}

type Notification struct {
	Revision  uint64
	EventType string
}

func (p *notifications) Publish(n Notification) {
	p.RLock()
	for ch := range p.subscribers {
		for {
			select {
			case ch <- n:
			case <-ch:
				continue
			}
			break
		}
	}
	p.RUnlock()
}

func (p *notifications) Subscribe() (<-chan Notification, func()) {
	ch := make(chan Notification, 1)
	p.Lock()
	p.subscribers[ch] = true
	p.Unlock()
	return ch, func() {
		p.Lock()
		delete(p.subscribers, ch)
		p.Unlock()
		close(ch)
	}
}
