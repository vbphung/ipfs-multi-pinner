package queue

import (
	"fmt"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	q := New[int]()

	done := make(chan struct{})
	go func() {
		offset := uint64(0)
		for {
			next := q.Sub(offset)
			if next != nil {
				fmt.Println(next.event)
				offset++

				if offset == 100 {
					done <- struct{}{}
					return
				}

				time.Sleep(1 * time.Second)
			}
		}
	}()

	for i := 0; i < 100; i++ {
		q.Pub(i, 1)
	}

	<-done
	close(done)
}
