package manager

import (
	"io"
	"sync"

	"github.com/jimmydrinkscoffee/easipfs/core"
)

type action struct {
	pin   *core.CID
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
