package subscribe

import (
	"context"
	"sync"
)

type FilterPublisher struct {
	mu  sync.Mutex
	f   func(*Message) bool
	pub Publisher
}

func NewFilterPublisher(pub Publisher) *FilterPublisher {
	return &FilterPublisher{
		f: func(*Message) bool {
			return false
		},
		pub: pub,
	}
}

func (p *FilterPublisher) SetFilter(f func(*Message) bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.f = f
}

func (p *FilterPublisher) Publish(ctx context.Context, message *Message) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.f(message) {
		p.pub.Publish(ctx, message)
	}
}
