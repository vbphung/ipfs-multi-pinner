package queue

import (
	"errors"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

type Msg[T any] struct {
	event T
	cons  uint32
}

type Queue[T any] interface {
	Pub(msg T, cons uint32)
	Sub() (*Consumer[T], error)
	Ack(consID string) error
}

type queue[T any] struct {
	cons map[string]*Consumer[T]
	msgs map[uint64]*Msg[T]
	h    uint64
	mu   sync.Mutex
}

func New[T any]() Queue[T] {
	return &queue[T]{
		cons: make(map[string]*Consumer[T]),
		msgs: make(map[uint64]*Msg[T]),
		h:    1,
	}
}

func (q *queue[T]) Pub(event T, cons uint32) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.h++
	q.msgs[q.h] = &Msg[T]{event, cons}

	for _, cons := range q.cons {
		if cons.status == consWait {
			cons.Sub <- q.msgs[q.h]
			cons.status = consBusy

			return
		}
	}
}

func (q *queue[T]) Sub() (*Consumer[T], error) {
	consID, err := uuid.NewV6()
	if err != nil {
		return nil, err
	}

	cons := &Consumer[T]{
		ID:     consID.String(),
		Sub:    make(chan *Msg[T]),
		offset: 1,
		status: consWait,
	}

	q.cons[cons.ID] = cons

	err = q.subNext(cons.ID)
	if err != nil {
		delete(q.cons, cons.ID)

		return nil, err
	}

	return cons, nil
}

func (q *queue[T]) Ack(consID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	cons, ok := q.cons[consID]
	if !ok {
		return errors.New(consID)
	}

	err := q.onConsAck(cons.offset + 1)
	if err != nil {
		return err
	}

	cons.offset++

	err = q.subNext(consID)
	if err != nil {
		return err
	}

	return nil
}

func (q *queue[T]) subNext(consID string) error {
	cons, ok := q.cons[consID]
	if !ok {
		return errors.New(consID)
	}

	for i := cons.offset + 1; i < q.h; i++ {
		if msg, ok := q.msgs[i]; ok {
			cons.Sub <- msg
			cons.status = consBusy

			return nil
		}
	}

	cons.status = consWait

	return nil
}

func (q *queue[T]) onConsAck(offset uint64) error {
	msg, ok := q.msgs[offset]
	if !ok {
		return errors.New(strconv.FormatUint(offset, 10))
	}

	msg.cons--
	if msg.cons == 0 {
		delete(q.msgs, offset)
	}

	return nil
}
