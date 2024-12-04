package manager

import (
	"errors"
	"io"
	"sync"

	"github.com/vbphung/ipfs-multi-pinner/core"
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
	ch  chan actRes
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

func (a *action) publish(r actRes) {
	a.ch <- r
}

func (a *action) unlock() {
	a.add.mu.Unlock()
}

func (a *action) done() {
	a.add.mu.Lock()
	close(a.ch)
}

func (a *action) res() (*core.CID, error) {
	var res actRes

	for r := range a.ch {
		if r.res != nil {
			res.res = r.res
		}

		if r.err != nil {
			res.err = errors.Join(res.err, r.err)
		}
	}

	if res.res != nil {
		return res.res, nil
	}

	return nil, res.err
}
