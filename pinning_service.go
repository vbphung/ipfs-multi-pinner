package easipfs

import (
	"context"
	"io"
)

type PinningService interface {
	Name() string
	Add(ctx context.Context, r io.Reader) (*CID, error)
	Pin(ctx context.Context, cid *CID) error
}
