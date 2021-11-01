package retranslator

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/app/repo"
	"github.com/BarchDif/stm-like-api/internal/app/sender"
	"github.com/BarchDif/stm-like-api/internal/app/workerpool"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	streaming "github.com/BarchDif/stm-like-api/internal/model"
	"github.com/BarchDif/stm-like-api/internal/testHelper"
	"github.com/golang/mock/gomock"
	"sync"
	"testing"
	"time"
)

var testData = struct {
	batch1 []streaming.LikeEvent
	batch2 []streaming.LikeEvent
}{
	batch1: []streaming.LikeEvent{
		{ID: 20},
		{ID: 21},
		{ID: 22},
	},
	batch2: []streaming.LikeEvent{
		{ID: 10},
		{ID: 11},
		{ID: 12},
	},
}

var cfg = Config{
	ChannelSize:    512,
	ConsumerCount:  2,
	ConsumeSize:    3,
	ConsumeTimeout: 100 * time.Millisecond,
	ProducerCount:  2,
}

func newConfig(repo repo.EventRepo, sender sender.EventSender, pool workerpool.WorkerPool) Config {
	cfg.Repo = repo
	cfg.Sender = sender
	cfg.Pool = pool

	return cfg
}

func createMocks(t *testing.T) (repo *mocks.MockEventRepo, sender *mocks.MockEventSender, pool *mocks.MockWorkerPool) {
	ctrl := gomock.NewController(t)
	repo = mocks.NewMockEventRepo(ctrl)
	sender = mocks.NewMockEventSender(ctrl)
	pool = mocks.NewMockWorkerPool(ctrl)

	return
}

func setupRetranslator(repo repo.EventRepo, sender sender.EventSender, pool workerpool.WorkerPool) Retranslator {
	return NewRetranslator(newConfig(repo, sender, pool))
}

func TestStart_SuccessFlow(t *testing.T) {
	subsetMatcher := testHelper.IsSubset(append(testData.batch1, testData.batch2...))
	repo, sender, pool := createMocks(t)
	retranslator := setupRetranslator(repo, sender, pool)

	wg := sync.WaitGroup{}
	wg.Add(2)

	repo.EXPECT().Lock(cfg.ConsumeSize).DoAndReturn(func(_ uint64) ([]streaming.LikeEvent, error) {
		defer wg.Done()

		return testData.batch1, nil
	})
	repo.EXPECT().Lock(cfg.ConsumeSize).DoAndReturn(func(_ uint64) ([]streaming.LikeEvent, error) {
		defer wg.Done()

		return testData.batch2, nil
	})

	sender.EXPECT().Send(subsetMatcher).Times(len(testData.batch1) + len(testData.batch2)).Return(nil)
	repo.EXPECT().Remove(subsetMatcher).MaxTimes(len(testData.batch1) + len(testData.batch2)).Return(nil)

	pool.EXPECT().Start(gomock.Any())
	pool.EXPECT().Submit(gomock.Any()).MaxTimes(len(testData.batch1) + len(testData.batch2)).DoAndReturn(func(f func() error) error {
		return f()
	})
	pool.EXPECT().StopWait().DoAndReturn(func() chan struct{} {
		stopped := make(chan struct{}, 1)
		stopped <- struct{}{}
		close(stopped)

		return stopped
	})

	retranslator.Start(context.Background())
	wg.Wait()
	<-retranslator.Stopped()
}
