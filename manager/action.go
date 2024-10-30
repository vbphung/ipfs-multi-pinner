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
	ch := make(chan actRes)

	go func() {
		defer close(ch)
		var (
			err     error
			success = false
		)
		for {
			r, ok := <-a.ch
			if !ok {
				break
			}

			if r.err == nil && !success {
				success = true
				ch <- actRes{r.res, nil}
			} else if err == nil {
				err = r.err
			}
		}

		ch <- actRes{nil, err}
	}()

	r := <-ch
	return r.res, r.err
}
