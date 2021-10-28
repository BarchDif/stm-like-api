package consumer_test

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/app/consumer"
	"github.com/BarchDif/stm-like-api/internal/model"
	"github.com/BarchDif/stm-like-api/internal/tests/fixture"
	"sync"
	"testing"
	"time"
)

const (
	consumerCount = 3
	batchSize     = 3
)

func startNewConsumer(t *testing.T) (consumer.Consumer, chan streaming.LikeEvent, map[uint64]int, *sync.Mutex) {
	testFixture := fixture.NewFixture(t)

	match := make(map[uint64]int)
	mu := new(sync.Mutex)

	repo := testFixture.LikeRepoLockMock(mu, match)
	resultChannel := make(chan streaming.LikeEvent, batchSize*consumerCount)

	consumer := consumer.NewDbConsumer(consumerCount, batchSize, time.Millisecond*10, repo, resultChannel)
	ctx, _ := context.WithCancel(context.Background())
	consumer.Start(ctx)

	return consumer, resultChannel, match, mu
}

// Checks if all generated events consumed once
func TestConsumer_EventsReceiving(t *testing.T) {
	consumer, resultChannel, match, mu := startNewConsumer(t)

	time.Sleep(time.Millisecond * 100)
	consumer.Cancel()

	if len(resultChannel) == 0 {
		t.Fail()

		return
	}

	for event := range resultChannel {
		mu.Lock()

		count, ok := match[event.ID]
		if !ok || count != 0 {
			t.Fail()

			return
		}

		match[event.ID]++
		mu.Unlock()
	}

	for _, value := range match {
		if value != 1 {
			t.Fail()

			return
		}
	}

	<-consumer.Cancel()
}
