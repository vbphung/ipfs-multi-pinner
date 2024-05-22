package easipfs

import (
	"context"
	"io"
)

type PinningService interface {
	Add(ctx context.Context, r io.Reader) (*CID, error)
	Pin(ctx context.Context, cid *CID) error
}
