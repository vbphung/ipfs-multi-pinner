package easipfs

import (
	"context"
	"io"
	"sync"

	"github.com/jimmydrinkscoffee/easipfs/queue"
	"github.com/sirupsen/logrus"
)

type action struct {
	pin   *CID
	add   io.Reader
	addMu sync.Mutex
}

func (a *action) tee() (io.Reader, bool) {
	a.addMu.Lock()

	if a.add == nil {
		return nil, false
	}

	res, next := teeReader(a.add)
	a.add = next

	return res, true
}

func (a *action) unlock() {
	a.addMu.Unlock()
}

type consumer struct {
	pn   PinService
	cons *queue.Consumer[*action]
}

type pinManager struct {
	q    *queue.Queue[*action]
	cons []*consumer
	log  *logrus.Logger
}

func newPinManager(log *logrus.Logger, pns ...PinService) (*pinManager, error) {
	q := queue.New[*action](log)

	pns = sliceFilter(pns, func(pn PinService) bool {
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

	return &pinManager{q, cons, log}, nil
}

func (m *pinManager) add(buf io.Reader) {
	m.q.Pub(&action{
		add: buf,
	})
}

func (m *pinManager) pin(cid *CID) {
	m.q.Pub(&action{
		pin: cid,
	})
}

func (m *pinManager) start() {
	for _, c := range m.cons {
		go func(cons *consumer) {
			for act := range cons.cons.Sub() {
				m.handleAction(cons.pn, act)
				cons.cons.Ack()
			}
		}(c)
	}
}

func (m *pinManager) handleAction(pn PinService, act *action) {
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
