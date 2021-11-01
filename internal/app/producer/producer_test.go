package producer

import (
	"context"
	"errors"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	"github.com/BarchDif/stm-like-api/internal/model"
	"github.com/BarchDif/stm-like-api/internal/testHelper"
	"github.com/golang/mock/gomock"
	"testing"
)

const (
	producerCount = 3
)

var testData = struct {
	batch []streaming.LikeEvent
}{
	batch: []streaming.LikeEvent{
		{ID: 10},
		{ID: 11},
		{ID: 12},
		{ID: 20},
		{ID: 21},
		{ID: 22},
	},
}

func createMocks(t *testing.T) (repo *mocks.MockEventRepo, sender *mocks.MockEventSender, pool *mocks.MockWorkerPool) {
	ctrl := gomock.NewController(t)
	repo = mocks.NewMockEventRepo(ctrl)
	sender = mocks.NewMockEventSender(ctrl)
	pool = mocks.NewMockWorkerPool(ctrl)

	return
}

func setupProducer(repo *mocks.MockEventRepo, sender *mocks.MockEventSender, pool *mocks.MockWorkerPool) Producer {
	events := make(chan streaming.LikeEvent, len(testData.batch))
	for _, event := range testData.batch {
		events <- event
	}
	close(events)

	producer := NewKafkaProducer(producerCount, repo, sender, events, pool)

	return producer
}

func TestProducer_SuccessfulSend(t *testing.T) {
	repo, sender, pool := createMocks(t)
	subsetMatcher := testHelper.IsSubset(testData.batch)
	producer := setupProducer(repo, sender, pool)

	sender.EXPECT().Send(subsetMatcher).Times(len(testData.batch)).Return(nil)
	pool.EXPECT().Submit(gomock.Any()).MaxTimes(len(testData.batch)).DoAndReturn(func(f func() error) error {
		return f()
	})
	repo.EXPECT().Remove(subsetMatcher).MaxTimes(len(testData.batch)).Return(nil)

	producer.Start(context.Background())
	<-producer.Cancel()
}

func TestProducer_FailedSend(t *testing.T) {
	repo, sender, pool := createMocks(t)
	subsetMatcher := testHelper.IsSubset(testData.batch)
	producer := setupProducer(repo, sender, pool)

	sender.EXPECT().Send(subsetMatcher).Times(len(testData.batch)).Return(errors.New("Send error"))
	pool.EXPECT().Submit(gomock.Any()).MaxTimes(len(testData.batch)).DoAndReturn(func(f func() error) error {
		return f()
	})
	repo.EXPECT().Unlock(subsetMatcher).MaxTimes(len(testData.batch)).Return(nil)

	producer.Start(context.Background())
	<-producer.Cancel()
}
