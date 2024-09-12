package manager

import (
	"io"
	"sync"

	"github.com/vbphung/easipfs/core"
)

type addAct struct {
	buf io.Reader
	mu  sync.Mutex
}

type pinAct core.CID

type actRes struct {
	res *core.CID
	err error
}

type action struct {
	add *addAct
	pin *pinAct
	ch  chan *actRes
}

func (a *action) check() (io.Reader, bool) {
	if a.add == nil {
		return nil, false
	}

	a.add.mu.Lock()

	res, next := teeReader(a.add.buf)
	a.add.buf = next

	return res, true
}

func (a *action) unlock() {
	a.add.mu.Unlock()
}

func (a *action) done() {
	a.add.mu.Lock()

	close(a.ch)
}

func (a *action) res() (*core.CID, error) {
	var err error
	for r := range a.ch {
		if r.err == nil {
			return r.res, nil
		}

		err = r.err
	}

	return nil, err
}
