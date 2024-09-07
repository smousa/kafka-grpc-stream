package subscribe

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

type FilterFunc func(*Message) bool

type FilterPublisher struct {
	mu  sync.Mutex
	f   FilterFunc
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

func (p *FilterPublisher) SetFilter(f FilterFunc) {
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

func FilterAnd(filterFuncs ...FilterFunc) FilterFunc {
	return func(m *Message) bool {
		for _, f := range filterFuncs {
			if !f(m) {
				return false
			}
		}

		return true
	}
}

func FilterKeys(keys []string) (FilterFunc, error) {
	pattern, err := glob.Compile(fmt.Sprintf("{%s}", strings.Join(keys, ",")))
	if err != nil {
		return nil, errors.Wrap(err, "invalid key glob")
	}

	return func(m *Message) bool {
		return pattern.Match(m.Key)
	}, nil
}

func FilterMaxAge(age string) (FilterFunc, error) {
	dur, err := time.ParseDuration(age)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse age duration")
	}

	return func(m *Message) bool {
		return time.Since(m.Timestamp) <= dur
	}, nil
}

func FilterMinOffset(minOffset int64) FilterFunc {
	return func(m *Message) bool {
		return m.Offset >= minOffset
	}
}
