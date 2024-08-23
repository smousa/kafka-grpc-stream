package subscribe

import (
	"context"
	"sync"
)

type Broadcast struct {
	mu   sync.Mutex
	pubs map[string]Publisher
}

func NewBroadcast() *Broadcast {
	return &Broadcast{pubs: make(map[string]Publisher)}
}

func (b *Broadcast) RegisterSession(sessionId string, pub Publisher) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.pubs[sessionId]; ok {
		panic("duplicate session id")
	}

	b.pubs[sessionId] = pub
}

func (b *Broadcast) UnregisterSession(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.pubs, id)
}

func (b *Broadcast) Publish(ctx context.Context, message *Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, pub := range b.pubs {
		pub.Publish(ctx, message)
	}
}
