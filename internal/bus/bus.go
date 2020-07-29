package bus

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	primary *Bus
	once    sync.Once

	ClosedErr  = errors.New("topic is closed")
	TimeoutErr = errors.New("push test timed out, no push received on time")
)

func init() {
	once.Do(func() {
		primary = New()
	})
}

func Primary() *Bus {
	return primary
}

func New() *Bus {
	return &Bus{
		topics: map[string]chan []byte{},
	}
}

type Bus struct {
	mu     sync.RWMutex
	topics map[string]chan []byte
}

func (b *Bus) createChan(topic string) chan []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	c, ok := b.topics[topic]
	if !ok {
		b.topics[topic] = make(chan []byte, 1)
		c = b.topics[topic]
	}

	return c
}

func (b *Bus) Publish(topic string, content []byte) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("topic is closed: %v", r)
		}
	}()

	b.mu.RLock()
	c, ok := b.topics[topic]
	b.mu.RUnlock()

	if !ok {
		c = b.createChan(topic)
	}

	select {
	case c <- content:
		return nil
	default:
		return errors.New("queue is full")
	}

}

func (b *Bus) Close(topic string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	c, ok := b.topics[topic]
	if !ok {
		return errors.New("does not exist")
	}
	delete(b.topics, topic)
	close(c)
	return nil
}

func (b *Bus) Next(topic string, timeout time.Duration) ([]byte, error) {
	b.mu.RLock()
	c, ok := b.topics[topic]
	b.mu.RUnlock()
	if !ok {
		c = b.createChan(topic)
	}

	select {
	case data, ok := <-c:
		if !ok {
			return nil, ClosedErr
		}
		return data, nil
	case <-time.After(timeout):
		return nil, TimeoutErr
	}
}
