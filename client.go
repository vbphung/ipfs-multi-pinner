package easipfs

import (
	"context"

	"github.com/jimmydrinkscoffee/easipfs/core"
	"github.com/jimmydrinkscoffee/easipfs/manager"
	"github.com/sirupsen/logrus"
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

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	clt, err := manager.New(log, pns...)
	if err != nil {
		return nil, err
	}

	clt.Start(context.Background())

	return clt, nil
}
