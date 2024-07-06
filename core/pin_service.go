package core

import (
	"context"
	"io"
)

type CID struct {
	Hash string
}

type PinService interface {
	Name() string
	Add(ctx context.Context, r io.Reader) (*CID, error)
	Pin(ctx context.Context, cid *CID) error
	Get(ctx context.Context, cid *CID) (io.Reader, error)
	ListCID(ctx context.Context) (<-chan *CID, error)
}
