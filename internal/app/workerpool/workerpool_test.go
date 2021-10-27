package workerpool

import (
	"context"
	"testing"
)

const workerCount = 2

func TestSimpleWorkerPool_StartStop(t *testing.T) {
	pool := New(workerCount)

	t.Run("Start", func(t *testing.T) {
		ctx, _ := context.WithCancel(context.Background())
		pool.Start(ctx)
	})

	t.Run("Cancel", func(t *testing.T) {
		<-pool.StopWait()
	})
}
