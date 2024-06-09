package queue

import "sync"

type Msg[T any] struct {
	event T
	cons  uint32
}

type Queue[T any] interface {
	Pub(msg T, cons uint32)
	Sub(prev uint64) *Msg[T]
}

type queue[T any] struct {
	msgs map[uint64]*Msg[T]
	h    uint64
	mu   sync.Mutex
}

func New[T any]() Queue[T] {
	return &queue[T]{
		msgs: make(map[uint64]*Msg[T]),
		h:    1,
	}
}

func (q *queue[T]) Pub(event T, cons uint32) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.h++
	q.msgs[q.h] = &Msg[T]{event, cons}
}

func (q *queue[T]) Sub(prev uint64) *Msg[T] {
	q.mu.Lock()
	defer q.mu.Unlock()

	if msg, ok := q.msgs[prev]; ok {
		msg.cons--
		if msg.cons == 0 {
			delete(q.msgs, prev)
		}
	}

	for i := prev + 1; i <= q.h; i++ {
		if msg, ok := q.msgs[i]; ok {
			return msg
		}
	}

	return nil
}
