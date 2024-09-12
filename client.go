package easipfs

import (
	"context"

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

	clt, err := manager.New(pns...)
	if err != nil {
		return nil, err
	}

	clt.Start(context.Background())

	return clt, nil
}
