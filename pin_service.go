package easipfs

import (
	"context"
	"io"
)

type PinService interface {
	Name() string
	Add(ctx context.Context, r io.Reader) (*CID, error)
	Pin(ctx context.Context, cid *CID) error
}
