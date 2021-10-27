package consumer

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	"github.com/BarchDif/stm-like-api/internal/model"
	"github.com/golang/mock/gomock"
	"math/rand"
	"testing"
	"time"
)

const (
	consumerCount = 3
	batchSize     = 5
)

func TestConsumer_StartCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockEventRepo(ctrl)
	resultChannel := make(chan streaming.LikeEvent)

	consumer := NewDbConsumer(consumerCount, batchSize, time.Millisecond*10, repo, resultChannel)

	t.Run("Start", func(t *testing.T) {
		ctx, _ := context.WithCancel(context.Background())
		consumer.Start(ctx)
	})

	t.Run("Cancel", func(t *testing.T) {
		<-consumer.Cancel()
	})
}

func TestConsumer_EventsReceiving(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockEventRepo(ctrl)
	resultChannel := make(chan streaming.LikeEvent)

	repo.EXPECT().Lock(gomock.Any()).AnyTimes().Do(func(n uint64) {
		for i := uint64(0); i < n; i++ {
			newEvent := streaming.LikeEvent{
				ID:     rand.Uint64(),
				Type:   streaming.Created,
				Status: streaming.Deferred,
				Entity: nil,
			}

			resultChannel <- newEvent
		}
	})

	consumer := NewDbConsumer(consumerCount, batchSize, time.Millisecond*10, repo, resultChannel)
	ctx, _ := context.WithCancel(context.Background())
	consumer.Start(ctx)

	for i := 0; i < batchSize*consumerCount; i++ {
		<-resultChannel
	}

	<-consumer.Cancel()
}
