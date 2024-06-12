package easipfs

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type testPinService struct {
	name  string
	delay time.Duration
}

func newTestPinService(name string, delay time.Duration) *testPinService {
	return &testPinService{name, delay}
}

func (t *testPinService) Add(ctx context.Context, r io.Reader) (*CID, error) {
	time.Sleep(t.delay)

	hash, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &CID{
		Hash: hash.String(),
	}, nil
}

func (t *testPinService) Pin(ctx context.Context, cid *CID) error {
	time.Sleep(t.delay)
	return nil
}

func (t *testPinService) Name() string {
	return t.name
}

func TestPinServices(t *testing.T) {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	pns := make([]PinService, 3)
	for i := range pns {
		name, err := uuid.NewV6()
		require.NoError(t, err)

		pns[i] = newTestPinService(name.String(), 500*time.Millisecond)
	}

	mng, err := newPinManager(log, pns...)
	require.NoError(t, err)

	mng.start()

	for i := 0; i < 2; i++ {
		r, err := os.Open(filePath)
		require.NoError(t, err)
		defer r.Close()

		mng.add(r)
	}

	for i := 0; i < 100; i++ {
		hash, err := uuid.NewV7()
		require.NoError(t, err)

		mng.pin(&CID{hash.String()})
	}

	time.Sleep(5 * time.Second)
}
