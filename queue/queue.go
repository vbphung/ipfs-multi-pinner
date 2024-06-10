package queue

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Msg[T any] struct {
	event T
	cons  uint32
}

type Queue[T any] struct {
	cons map[string]*Consumer[T]
	msgs map[uint64]*Msg[T]
	h    uint64
	mu   sync.Mutex
	log  *logrus.Logger
}

func New[T any](log *logrus.Logger) *Queue[T] {
	return &Queue[T]{
		cons: make(map[string]*Consumer[T]),
		msgs: make(map[uint64]*Msg[T]),
		h:    0,
		log:  log,
	}
}

func (q *Queue[T]) Pub(event T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.h++
	q.msgs[q.h] = &Msg[T]{event, uint32(len(q.cons))}

	for _, cons := range q.cons {
		if cons.status == consWait {
			q.subNext(cons.consID)

			return
		}
	}
}

func (q *Queue[T]) Sub() (*Consumer[T], error) {
	consID, err := uuid.NewV6()
	if err != nil {
		return nil, err
	}

	cons := newConsumer[T](consID.String(), q)
	q.cons[cons.consID] = cons

	err = q.subNext(cons.consID)
	if err != nil {
		delete(q.cons, cons.consID)

		return nil, err
	}

	return cons, nil
}

func (q *Queue[T]) ack(consID string) error {
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

	cons.incOffset()

	go func() {
		if err = q.subNext(consID); err != nil {
			fmt.Println(err)
		}
	}()

	return nil
}

func (q *Queue[T]) subNext(consID string) error {
	cons, ok := q.cons[consID]
	if !ok {
		return errors.New(consID)
	}

	for i := cons.offset + 1; i < q.h; i++ {
		if msg, ok := q.msgs[i]; ok {
			cons.sub(msg)

			return nil
		}
	}

	return nil
}

func (q *Queue[T]) onConsAck(offset uint64) error {
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
