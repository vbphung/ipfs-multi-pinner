package easipfs

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	clt, err := NewClient(&Config{
		IpfsUrl:    "http://localhost:5001",
		CIDVersion: 0,
	})
	require.NoError(t, err)

	r, err := os.Open("test.jpg")
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

	require.Contains(t, cids, added.Hash)
}
