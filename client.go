package ipfsmultipinner

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/vbphung/ipfs-multi-pinner/core"
	"github.com/vbphung/ipfs-multi-pinner/manager"
)

type Client core.PinService

func NewClient(selfs []*core.SelfHostedConf, pns []core.PinService) (Client, error) {
	for _, conf := range selfs {
		pn, err := core.NewSelfHostedService(conf)
		if err != nil {
			return nil, fmt.Errorf("connect self-hosted: %v", err)
		}

		pns = append(pns, pn)
	}

	if len(pns) == 0 {
		return nil, errors.New("there are no services")
	}

	m := make(map[string]core.PinService)
	for _, pn := range pns {
		m[pn.Name()] = pn
	}

	pns = slices.Collect(maps.Values(m))

	clt, err := manager.New(pns...)
	if err != nil {
		return nil, fmt.Errorf("create manager: %v", err)
	}

	clt.Start(context.Background())

	return clt, nil
}
