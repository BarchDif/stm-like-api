package consumer_test

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/app/consumer"
	"github.com/BarchDif/stm-like-api/internal/app/repo"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	"github.com/BarchDif/stm-like-api/internal/model"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

const (
	consumerCount  = 2
	batchSize      = 3
	consumeTimeout = time.Millisecond * 100
)

var testData = struct {
	batch1 []streaming.LikeEvent
	batch2 []streaming.LikeEvent
}{
	batch1: []streaming.LikeEvent{
		{ID: 10},
		{ID: 11},
		{ID: 12},
	},
	batch2: []streaming.LikeEvent{
		{ID: 20},
		{ID: 21},
		{ID: 22},
	},
}

func oneOfTestData(event streaming.LikeEvent) bool {
	for _, testEvent := range testData.batch1 {
		if event == testEvent {
			return true
		}
	}

	for _, testEvent := range testData.batch2 {
		if event == testEvent {
			return true
		}
	}

	return false
}

func createRepoMock(t *testing.T) *mocks.MockEventRepo {
	ctrl := gomock.NewController(t)
	repoMock := mocks.NewMockEventRepo(ctrl)

	return repoMock
}

func setupConsumer(repoMock repo.EventRepo) (consumer.Consumer, chan streaming.LikeEvent) {
	events := make(chan streaming.LikeEvent, consumerCount*batchSize)
	consumer := consumer.NewDbConsumer(consumerCount, batchSize, consumeTimeout, repoMock, events)

	return consumer, events
}

func TestConsumer_EventsReceiving(t *testing.T) {
	repoMock := createRepoMock(t)

	repoMock.EXPECT().Lock(uint64(batchSize)).Return(testData.batch1, nil)
	repoMock.EXPECT().Lock(uint64(batchSize)).Return(testData.batch2, nil)

	consumer, result := setupConsumer(repoMock)

	consumer.Start(context.Background())
	time.Sleep(consumeTimeout)
	<-consumer.Cancel()

	for event := range result {
		if !oneOfTestData(event) {
			t.Fail()
		}
	}
}
