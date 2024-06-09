package queue

import "sync"

type consumerStatus uint16

const (
	consWait consumerStatus = 0
	consBusy consumerStatus = 1
)

type Consumer[T any] struct {
	consID string
	ch     chan *Msg[T]
	offset uint64
	status consumerStatus
	q      *Queue[T]
	mu     sync.Mutex
}

func (c *Consumer[T]) Ack() error {
	return c.q.ack(c.consID)
}

func (c *Consumer[T]) Sub() <-chan *Msg[T] {
	return c.ch
}

func (c *Consumer[T]) incOffset() {
	defer c.mu.Unlock()
	c.mu.Lock()

	c.offset++
	c.status = consWait
}

func (c *Consumer[T]) sub(msg *Msg[T]) {
	defer c.mu.Unlock()
	c.mu.Lock()

	c.ch <- msg
	c.status = consBusy
}
