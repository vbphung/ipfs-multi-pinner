package easipfs

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
)

const (
	actionQueueSize     = 100
	defaultMaxReqAtOnce = 3
)

type action struct {
	pin *CID
	add io.Reader
}

type actionQueue struct {
	pn     PinningService
	ch     chan *action
	splits int
}

type pinningManager struct {
	pns []*actionQueue
	log *logrus.Logger
}

func newPinningManager(log *logrus.Logger, pns ...PinningService) *pinningManager {
	queues := make([]*actionQueue, len(pns))
	for i, pn := range pns {
		queues[i] = &actionQueue{
			pn:     pn,
			ch:     make(chan *action, actionQueueSize),
			splits: defaultMaxReqAtOnce,
		}
	}

	return &pinningManager{
		pns: queues,
		log: log,
	}
}

func (m *pinningManager) add(buf io.Reader) {
	cur := buf

	for _, pn := range m.pns {
		var next io.Reader
		next, cur = teeIoReader(cur)

		pn.ch <- &action{
			add: next,
		}
	}
}

func (m *pinningManager) pin(cid *CID) {
	for _, pn := range m.pns {
		pn.ch <- &action{
			pin: cid,
		}
	}
}

func (m *pinningManager) do() {
	for _, q := range m.pns {
		for i := 0; i < q.splits; i++ {
			go q.do(m.log)
		}
	}
}

func (q *actionQueue) do(log *logrus.Logger) {
	for {
		act, ok := <-q.ch
		if !ok {
			return
		}

		if act.add != nil {
			cid, err := q.pn.Add(context.Background(), act.add)
			if err != nil {
				log.Errorln(q.pn.Name(), err)
				continue
			}

			log.Infoln(q.pn.Name(), cid.Hash)
			continue
		}

		err := q.pn.Pin(context.Background(), act.pin)
		if err != nil {
			log.Errorln(q.pn.Name(), err)
			continue
		}

		log.Infoln(q.pn.Name(), act.pin.Hash)
	}
}
