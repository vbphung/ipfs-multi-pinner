package ipfsmultipinner

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vbphung/ipfs-multi-pinner/core"
)

type testPinningService struct {
	name string
	m    map[string][]byte
	mu   sync.RWMutex
}

func newTestPinningService(name string) core.PinService {
	return &testPinningService{
		name: name,
		m:    make(map[string][]byte),
	}
}

func (t *testPinningService) Add(ctx context.Context, r io.Reader) (*core.CID, error) {
	h := sha256.New()
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	h.Write(data)
	hash := hex.EncodeToString(h.Sum(nil))

	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.m[hash]; !ok {
		t.m[hash] = data
	}

	return &core.CID{
		Hash: hash,
	}, nil
}

func (t *testPinningService) Get(ctx context.Context, cid *core.CID) (io.Reader, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	data, ok := t.m[cid.Hash]
	if !ok {
		return nil, sql.ErrNoRows
	}

	return bytes.NewReader(data), nil
}

func (t *testPinningService) ListCID(ctx context.Context) (<-chan *core.CID, error) {
	t.mu.RLock()

	ch := make(chan *core.CID)
	go func() {
		defer t.mu.RUnlock()
		defer close(ch)
		for k := range t.m {
			ch <- &core.CID{Hash: k}
		}
	}()

	return ch, nil
}

func (t *testPinningService) Name() string {
	return t.name
}

func (t *testPinningService) Pin(ctx context.Context, cid *core.CID) error {
	return nil
}

func TestMain(t *testing.T) {
	s := make([]core.PinService, 3)
	for i := range 3 {
		s[i] = newTestPinningService(genString(t, 12))
	}
	clt, err := NewClient([]*core.SelfHostedConf{}, s)
	require.NoError(t, err)

	var (
		ctx  = context.Background()
		cids = make(map[string]bool)
	)
	for range 100 {
		fmt.Println()

		data := []byte(genString(t, 1024))
		for range 5 {
			cid, err := clt.Add(ctx, bytes.NewReader(data))
			require.NoError(t, err)
			cids[cid.Hash] = true
		}
		time.Sleep(100 * time.Millisecond)
	}

	ls, err := clt.ListCID(ctx)
	require.NoError(t, err)
	for cid := range ls {
		delete(cids, cid.Hash)
	}

	require.Empty(t, cids)
}

func genString(t *testing.T, n int) string {
	bytes := make([]byte, n)

	_, err := rand.Read(bytes)
	require.NoError(t, err)

	return hex.EncodeToString(bytes)
}
