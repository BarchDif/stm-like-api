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

func TestConsumer_EventsReceiving(t *testing.T) {
	testFixture := fixture.NewFixture(t)
	resultChannel := make(chan streaming.LikeEvent, batchSize*consumerCount)

	match := make(map[uint64]int)
	mu := new(sync.Mutex)

	repo := testFixture.LikeRepoMock(mu, match)

	consumer := consumer.NewDbConsumer(consumerCount, batchSize, time.Millisecond*10, repo, resultChannel)
	ctx, cancel := context.WithCancel(context.Background())

	consumer.Start(ctx)
	time.Sleep(time.Millisecond * 100)
	cancel()

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

	<-consumer.Cancel()
}
