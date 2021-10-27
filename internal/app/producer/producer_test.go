package producer

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	"github.com/BarchDif/stm-like-api/internal/model"
	"github.com/golang/mock/gomock"
	"sync/atomic"
	"testing"
)

const (
	producerCount = 3
	batchSize     = 5
)

func TestProducer_StartCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockEventRepo(ctrl)
	sender := mocks.NewMockEventSender(ctrl)
	events := make(chan streaming.LikeEvent)
	pool := mocks.NewMockWorkerPool(ctrl)

	producer := NewKafkaProducer(producerCount, repo, sender, events, pool)

	t.Run("Start", func(t *testing.T) {
		ctx, _ := context.WithCancel(context.Background())
		producer.Start(ctx)
	})

	t.Run("Cancel", func(t *testing.T) {
		close(events)
		<-producer.Cancel()
	})
}

func TestProducer_EventSendingSuccessful(t *testing.T) {
	ctrl := gomock.NewController(t)
	sender := mocks.NewMockEventSender(ctrl)
	repo := mocks.NewMockEventRepo(ctrl)
	events := make(chan streaming.LikeEvent, batchSize*producerCount)
	pool := mocks.NewMockWorkerPool(ctrl)

	eventsCount := int32(0)
	sender.EXPECT().Send(gomock.Any()).AnyTimes().DoAndReturn(func(event *streaming.LikeEvent) error {
		if event.Type == streaming.Created {
			atomic.AddInt32(&eventsCount, 1)
		}

		return nil
	})
	pool.EXPECT().Submit(gomock.Any()).AnyTimes()

	producer := NewKafkaProducer(producerCount, repo, sender, events, pool)
	ctx, _ := context.WithCancel(context.Background())

	t.Run("Send successful", func(t *testing.T) {
		for i := uint64(0); i < batchSize*producerCount; i++ {
			events <- streaming.LikeEvent{
				ID:   i,
				Type: streaming.Created,
			}
		}
		producer.Start(ctx)

		close(events)
		<-producer.Cancel()
		if eventsCount != batchSize*producerCount {
			t.Fail()
		}
	})
}
