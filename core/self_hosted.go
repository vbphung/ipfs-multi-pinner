package core

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/ipfs/kubo/core/coreiface/options"
)

type SelfHostedConf struct {
	IpfsUrl    string
	CIDVersion int
}

type selfHostedService struct {
	api  *rpc.HttpApi
	conf *SelfHostedConf
}

func NewSelfHostedService(conf *SelfHostedConf) (PinService, error) {
	api, err := rpc.NewURLApiWithClient(conf.IpfsUrl, &http.Client{})
	if err != nil {
		return nil, err
	}

	return &selfHostedService{api, conf}, nil
}

func (s *selfHostedService) Add(ctx context.Context, r io.Reader) (*CID, error) {
	added, err := s.api.Unixfs().Add(ctx, files.NewReaderFile(r), func(uas *options.UnixfsAddSettings) error {
		uas.CidVersion = s.conf.CIDVersion
		uas.Pin = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &CID{
		Hash: added.RootCid().String(),
	}, nil
}

func (s *selfHostedService) Get(ctx context.Context, req *CID) (io.Reader, error) {
	cid, err := cid.Decode(req.Hash)
	if err != nil {
		return nil, err
	}

	nd, err := s.api.Unixfs().Get(ctx, path.FromCid(cid))
	if err != nil {
		return nil, err
	}

	f, ok := nd.(files.File)
	if !ok {
		return nil, os.ErrNotExist
	}

	return f, nil
}

func (s *selfHostedService) ListCID(ctx context.Context) (<-chan *CID, error) {
	ch, err := s.api.Pin().Ls(ctx)
	if err != nil {
		return nil, err
	}

	res := make(chan *CID)
	go func(ch <-chan iface.Pin, res chan<- *CID) {
		defer close(res)

		for pin := range ch {
			res <- &CID{
				Hash: pin.Path().RootCid().String(),
			}
		}
	}(ch, res)

	return res, nil
}

func (s *selfHostedService) Name() string {
	return s.conf.IpfsUrl
}

func (s *selfHostedService) Pin(ctx context.Context, req *CID) error {
	cid, err := cid.Decode(req.Hash)
	if err != nil {
		return err
	}

	return s.api.Pin().Add(ctx, path.FromCid(cid))
}
