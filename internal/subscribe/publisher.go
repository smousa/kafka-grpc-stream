package subscribe

import "context"

type Publisher interface {
	Publish(ctx context.Context, message *Message)
}

type Publishers []Publisher

func (pubs Publishers) Publish(ctx context.Context, message *Message) {
	for _, pub := range pubs {
		pub.Publish(ctx, message)
	}
}
