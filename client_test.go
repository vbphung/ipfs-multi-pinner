package easipfs

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vbphung/easipfs/core"
)

const (
	url      = "http://localhost:5001"
	filePath = "test.jpg"
)

func TestAll(t *testing.T) {
	clt, err := NewClient([]*core.SelfHostedConf{{
		IpfsUrl:    url,
		CIDVersion: 1,
	}}, []core.PinService{})
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
