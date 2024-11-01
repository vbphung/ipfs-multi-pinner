package manager

import (
	"context"
	"io"

	"github.com/rs/zerolog/log"
	"github.com/vbphung/ipfs-multi-pinner/core"
	"github.com/vbphung/ipfs-multi-pinner/queue"
)

type consumer struct {
	pn   core.PinService
	cons *queue.Consumer[action]
}

type Manager interface {
	core.PinService

	Start(ctx context.Context)
}

type manager struct {
	q    *queue.Queue[action]
	cons []*consumer
}

func New(pns ...core.PinService) (Manager, error) {
	q := queue.New(func(a action) {
		a.done()
	})

	pns = sliceFilter(pns, func(pn core.PinService) bool {
		return pn != nil
	})
	cons := make([]*consumer, len(pns))
	for i, pn := range pns {
		sub, err := q.Sub()
		if err != nil {
			return nil, err
		}

		cons[i] = &consumer{
			pn:   pn,
			cons: sub,
		}
	}

	return &manager{q, cons}, nil
}

func (m *manager) Add(ctx context.Context, r io.Reader) (*core.CID, error) {
	act := action{
		add: &addAct{
			buf: r,
		},
		ch: make(chan actRes),
	}
	m.q.Pub(act)

	return act.res()
}

func (m *manager) Get(ctx context.Context, cid *core.CID) (io.Reader, error) {
	var err error
	for _, c := range m.cons {
		r, err := c.pn.Get(ctx, cid)
		if err == nil && r != nil {
			return r, nil
		}
	}

	return nil, err
}

func (m *manager) ListCID(ctx context.Context) (<-chan *core.CID, error) {
	var err error
	for _, c := range m.cons {
		ch, err := c.pn.ListCID(ctx)
		if err == nil && ch != nil {
			return ch, nil
		}
	}

	return nil, err
}

func (m *manager) Name() string {
	return m.cons[0].pn.Name()
}

func (m *manager) Pin(ctx context.Context, cid *core.CID) error {
	act := action{
		pin: (*pinAct)(cid),
		ch:  make(chan actRes),
	}
	m.q.Pub(act)

	_, err := act.res()
	if err != nil {
		return err
	}

	return nil
}

func (m *manager) Start(ctx context.Context) {
	for _, c := range m.cons {
		go func(cons *consumer) {
			for act := range cons.cons.Sub() {
				m.handleAction(cons.pn, act)
				cons.cons.Ack()
			}
		}(c)
	}
}

func (m *manager) handleAction(pn core.PinService, act action) {
	if r, ok := act.check(); ok {
		defer act.unlock()

		cid, err := pn.Add(context.Background(), r)
		if err != nil {
			log.Err(err).Msg(pn.Name())
			act.publish(actRes{err: err})

			return
		}

		log.Info().Str("name", pn.Name()).Str("hash", cid.Hash).Send()
		act.publish(actRes{res: cid})
	} else {
		err := pn.Pin(context.Background(), (*core.CID)(act.pin))
		if err != nil {
			log.Err(err).Msg(pn.Name())
			act.publish(actRes{err: err})

			return
		}

		log.Info().Str("name", pn.Name()).Str("hash", act.pin.Hash).Send()
		act.publish(actRes{res: (*core.CID)(act.pin)})
	}
}
