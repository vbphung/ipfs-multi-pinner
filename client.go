package easipfs

import (
	"context"
	"maps"
	"slices"

	"github.com/vbphung/easipfs/core"
	"github.com/vbphung/easipfs/manager"
)

type Client core.PinService

func NewClient(selfs []*core.SelfHostedConf, pns []core.PinService) (Client, error) {
	for _, conf := range selfs {
		pn, err := core.NewSelfHostedService(conf)
		if err != nil {
			return nil, err
		}

		pns = append(pns, pn)
	}

	m := make(map[string]core.PinService)
	for _, pn := range pns {
		m[pn.Name()] = pn
	}

	pns = slices.Collect(maps.Values(m))

	clt, err := manager.New(pns...)
	if err != nil {
		return nil, err
	}

	clt.Start(context.Background())

	return clt, nil
}
