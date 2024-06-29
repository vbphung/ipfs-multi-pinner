package manager

import (
	"io"
	"sync"

	"github.com/jimmydrinkscoffee/easipfs/core"
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
	res chan *actRes
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
