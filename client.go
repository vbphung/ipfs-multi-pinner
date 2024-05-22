package easipfs

import (
	"bufio"
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
	"github.com/multiformats/go-multihash"
)

type CID struct {
	Hash string
}

type Client interface {
	Add(ctx context.Context, r io.Reader) (*CID, error)
	Pin(ctx context.Context, cid *CID) error
	Get(ctx context.Context, cid *CID) (io.Reader, error)
	ListCID(ctx context.Context) (<-chan *CID, error)
}

type clt struct {
	api  *rpc.HttpApi
	pns  []PinningService
	conf *Config
}

func NewClient(conf *Config, pns []PinningService) (Client, error) {
	api, err := rpc.NewURLApiWithClient(conf.IpfsUrl, &http.Client{})
	if err != nil {
		return nil, err
	}

	return &clt{api, pns, conf}, nil
}

func (c *clt) Add(ctx context.Context, r io.Reader) (*CID, error) {
	buf := bufio.NewReader(r)

	added, err := c.api.Unixfs().Add(ctx, files.NewReaderFile(buf), func(uas *options.UnixfsAddSettings) error {
		uas.CidVersion = c.conf.CIDVersion
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &CID{
		Hash: added.RootCid().String(),
	}, nil
}

func (c *clt) Get(ctx context.Context, req *CID) (io.Reader, error) {
	mh, err := multihash.FromB58String(req.Hash)
	if err != nil {
		return nil, err
	}

	nd, err := c.api.Unixfs().Get(ctx, path.FromCid(cid.NewCidV0(mh)))
	if err != nil {
		return nil, err
	}

	f, ok := nd.(files.File)
	if !ok {
		return nil, os.ErrNotExist
	}

	return f, nil
}

func (c *clt) ListCID(ctx context.Context) (<-chan *CID, error) {
	ch, err := c.api.Pin().Ls(ctx)
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

func (c *clt) Pin(ctx context.Context, req *CID) error {
	mh, err := multihash.FromB58String(req.Hash)
	if err != nil {
		return err
	}

	return c.api.Pin().Add(ctx, path.FromCid(cid.NewCidV0(mh)))
}
