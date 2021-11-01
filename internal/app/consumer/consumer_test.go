package consumer

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/app/repo"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	"github.com/BarchDif/stm-like-api/internal/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

func createRepoMock(t *testing.T) *mocks.MockEventRepo {
	ctrl := gomock.NewController(t)
	repoMock := mocks.NewMockEventRepo(ctrl)

	return repoMock
}

func setupConsumer(repoMock repo.EventRepo) (Consumer, chan streaming.LikeEvent) {
	events := make(chan streaming.LikeEvent, consumerCount*batchSize)
	consumer := NewDbConsumer(consumerCount, batchSize, consumeTimeout, repoMock, events)

	return consumer, events
}

func chanToSlice(in chan streaming.LikeEvent) []streaming.LikeEvent {
	result := make([]streaming.LikeEvent, 0)

	for event := range in {
		result = append(result, event)
	}

	return result
}

func TestConsumer_AllEventsConsumed(t *testing.T) {
	a := assert.New(t)
	repoMock := createRepoMock(t)
	consumer, result := setupConsumer(repoMock)

	repoMock.EXPECT().Lock(uint64(batchSize)).Return(testData.batch1, nil)
	repoMock.EXPECT().Lock(uint64(batchSize)).Return(testData.batch2, nil)

	ctx, _ := context.WithTimeout(context.Background(), consumeTimeout)
	consumer.Start(ctx)

	a.ElementsMatch(append(testData.batch1, testData.batch2...), chanToSlice(result))
}
