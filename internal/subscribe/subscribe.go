package subscribe

import (
	"context"
	"time"
)

type Header struct {
	Key   string
	Value string
}

type Message struct {
	Key       string
	Value     []byte
	Headers   []Header
	Timestamp time.Time
	Topic     string
	Partition int32
	Offset    int64
}

type Publisher interface {
	Publish(ctx context.Context, message *Message)
}
