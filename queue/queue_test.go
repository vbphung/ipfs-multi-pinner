package queue

import (
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	q := New(func(n int) {
		log.Warn().Int("offset", n).Msg("message removed")
	})

	cons, err := q.Sub()
	require.NoError(t, err)

	go func(cons *Consumer[int]) {
		for {
			msg := <-cons.Sub()

			time.Sleep(500 * time.Millisecond)
			log.Info().Int("offset", msg).Msg("message processed")

			assert.NoError(t, cons.Ack())
		}
	}(cons)

	for i := 0; i < 100; i++ {
		q.Pub(i)
	}

	time.Sleep(3 * time.Second)
}
