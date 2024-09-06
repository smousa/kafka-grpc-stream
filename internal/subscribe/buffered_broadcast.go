package subscribe

import (
	"context"
	"sync"
	"time"

	"github.com/smousa/kafka-grpc-stream/internal/metrics"
)

type session struct {
	head  int
	empty bool
}

type BufferedBroadcast struct {
	mu       sync.RWMutex
	buf      []*Message
	ready    chan struct{}
	tail     int
	size     int
	limit    int
	sessions map[*session]struct{}
}

func NewBufferedBroadcast(n int) *BufferedBroadcast {
	return &BufferedBroadcast{
		buf:      make([]*Message, n),
		ready:    make(chan struct{}),
		limit:    n,
		sessions: make(map[*session]struct{}),
	}
}

func (bb *BufferedBroadcast) Publish(ctx context.Context, msg *Message) {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	// notify that a message is added
	defer func() {
		close(bb.ready)
		bb.ready = make(chan struct{})
	}()

	// update the current buffer fill size
	if bb.size < bb.limit {
		bb.size += 1
	} else {
		// report the age of the evicted message
		metrics.BufferedBroadcastEvictionTimestampSeconds.Set(
			float64(bb.buf[bb.tail].Timestamp.UnixMilli()) / float64(time.Second.Milliseconds()),
		)
	}

	// add the message to the buffer
	bb.buf[bb.tail] = msg

	// get the next tail
	tail := (bb.tail + 1) % bb.limit

	// update all the sessions
	for sess := range bb.sessions {
		if sess.head == bb.tail && !sess.empty {
			// the buffer is full, so move the head to indicate that the record
			// has been evicted.
			sess.head = tail

			// report the dropped message
			metrics.BufferedBroadcastDroppedMessagesTotal.Inc()
		}

		sess.empty = false
	}

	// update the tail
	bb.tail = tail
}

func (bb *BufferedBroadcast) RegisterSession(ctx context.Context, pub Publisher) {
	// get the session
	sess := func() *session {
		bb.mu.Lock()
		defer bb.mu.Unlock()

		sess := &session{
			empty: bb.size == 0,
		}

		// set the head pointer if the buffer is full
		if bb.size == bb.limit {
			sess.head = bb.tail
		}

		// add the session to the map
		bb.sessions[sess] = struct{}{}

		return sess
	}()

	// remove session when done
	defer func() {
		bb.mu.Lock()
		defer bb.mu.Unlock()
		delete(bb.sessions, sess)
	}()

	// publish whatever is in the buffer
	done := ctx.Done()

	for {
		bb.mu.RLock()

		if sess.empty {
			// buffer is empty, so wait for buffer
			ready := bb.ready
			bb.mu.RUnlock()
			select {
			case <-ready:
			case <-done:
				return
			}
			bb.mu.RLock()
		}

		// retrieve the next message from the buffer
		msg := bb.buf[sess.head]
		sess.head = (sess.head + 1) % bb.limit
		sess.empty = sess.head == bb.tail
		bb.mu.RUnlock()

		// publish the message or exit
		select {
		case <-done:
			return
		default:
			pub.Publish(ctx, msg)
		}
	}
}
