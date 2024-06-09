package queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	q := New[int]()

	cons, err := q.Sub()
	require.NoError(t, err)

	go func(cons *Consumer[int]) {
		for {
			msg := <-cons.Sub()

			time.Sleep(500 * time.Millisecond)
			fmt.Println(msg.event)

			assert.NoError(t, cons.Ack())
		}
	}(cons)

	for i := 0; i < 100; i++ {
		q.Pub(i)
	}

	time.Sleep(3 * time.Second)
}
