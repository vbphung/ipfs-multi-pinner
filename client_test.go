package easipfs

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	url      = "http://localhost:5001"
	filePath = "test.jpg"
)

type pseudoPinService struct {
	Client
	name string
}

func newPseudoPinService(clt Client, name string) *pseudoPinService {
	return &pseudoPinService{clt, name}
}

func (p *pseudoPinService) Name() string {
	return p.name
}

func TestAll(t *testing.T) {
	clt, err := NewClient(&Config{
		IpfsUrl:    url,
		CIDVersion: 0,
	})
	require.NoError(t, err)

	pns := make([]PinService, 3)
	for i := range pns {
		name, err := uuid.NewV7()
		require.NoError(t, err)

		pns[i] = newPseudoPinService(clt, name.String())
	}

	clt, err = NewClient(&Config{
		IpfsUrl:    url,
		CIDVersion: 0,
	}, pns...)
	require.NoError(t, err)

	r, err := os.Open(filePath)
	require.NoError(t, err)

	ctx := context.Background()

	added, err := clt.Add(ctx, r)
	require.NoError(t, err)
	fmt.Println(added.Hash)

	time.Sleep(5 * time.Second)

	ls, err := clt.ListCID(ctx)
	require.NoError(t, err)

	var cids []string
	for cid := range ls {
		cids = append(cids, cid.Hash)
	}

	require.Contains(t, cids, added.Hash)
}

func TestAllV1(t *testing.T) {
	clt, err := NewClient(&Config{
		IpfsUrl:    url,
		CIDVersion: 1,
	})
	require.NoError(t, err)

	r, err := os.Open(filePath)
	require.NoError(t, err)

	ctx := context.Background()

	added, err := clt.Add(ctx, r)
	require.NoError(t, err)
	fmt.Println(added.Hash)

	ls, err := clt.ListCID(ctx)
	require.NoError(t, err)

	var cids []string
	for cid := range ls {
		cids = append(cids, cid.Hash)
	}

	fmt.Println(len(cids), cids)

	require.Contains(t, cids, added.Hash)
}
