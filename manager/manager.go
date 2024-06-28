package manager

import (
	"context"
	"io"

	"github.com/jimmydrinkscoffee/easipfs/core"
	"github.com/jimmydrinkscoffee/easipfs/queue"
	"github.com/sirupsen/logrus"
)

type consumer struct {
	pn   core.PinService
	cons *queue.Consumer[*action]
}

type manager struct {
	q    *queue.Queue[*action]
	cons []*consumer
	log  *logrus.Logger
}

func New(log *logrus.Logger, pns ...core.PinService) (core.PinService, error) {
	q := queue.New[*action](log)

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

	return &manager{q, cons, log}, nil
}

func (m *manager) Add(ctx context.Context, r io.Reader) (*core.CID, error) {
	panic("unimplemented")
}

func (m *manager) Get(ctx context.Context, cid *core.CID) (io.Reader, error) {
	panic("unimplemented")
}

func (m *manager) ListCID(ctx context.Context) (<-chan *core.CID, error) {
	panic("unimplemented")
}

func (m *manager) Name() string {
	panic("unimplemented")
}

func (m *manager) Pin(ctx context.Context, cid *core.CID) error {
	panic("unimplemented")
}

func (m *manager) add(buf io.Reader) {
	m.q.Pub(&action{
		add: buf,
	})
}

func (m *manager) pin(cid *core.CID) {
	m.q.Pub(&action{
		pin: cid,
	})
}

func (m *manager) start() {
	for _, c := range m.cons {
		go func(cons *consumer) {
			for act := range cons.cons.Sub() {
				m.handleAction(cons.pn, act)
				cons.cons.Ack()
			}
		}(c)
	}
}

func (m *manager) handleAction(pn core.PinService, act *action) {
	if r, ok := act.tee(); ok {
		defer act.unlock()

		cid, err := pn.Add(context.Background(), r)
		if err != nil {
			m.log.Errorln(pn.Name(), err)
			return
		}

		m.log.Infoln(pn.Name(), cid.Hash)
	} else {
		err := pn.Pin(context.Background(), act.pin)
		if err != nil {
			m.log.Errorln(pn.Name(), err)
			return
		}

		m.log.Infoln(pn.Name(), act.pin.Hash)
	}
}
