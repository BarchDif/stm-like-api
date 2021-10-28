package retranslator

import (
	"context"
	"errors"
	"github.com/BarchDif/stm-like-api/internal/app/repo"
	retranslator2 "github.com/BarchDif/stm-like-api/internal/app/retranslator"
	"github.com/BarchDif/stm-like-api/internal/app/sender"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	streaming "github.com/BarchDif/stm-like-api/internal/model"
	"github.com/BarchDif/stm-like-api/internal/tests/fixture"
	"github.com/BarchDif/stm-like-api/internal/tests/testdata"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func setup(t *testing.T) (repo.EventRepo, sender.EventSender, chan streaming.LikeEvent) {
	testFixture := fixture.NewFixture(t)

	removed := make(chan streaming.LikeEvent, len(testdata.Events))

	repo := testFixture.LikeRepoMock(func(eventRepo *mocks.MockEventRepo) {
		lockMu := new(sync.Mutex)
		offset := 0
		eventRepo.
			EXPECT().
			Lock(gomock.AssignableToTypeOf(uint64(0))).
			AnyTimes().
			DoAndReturn(func(n uint64) ([]streaming.LikeEvent, error) {
				lockMu.Lock()
				defer lockMu.Unlock()

				end := offset + int(n)

				if offset >= len(testdata.Events) {
					return nil, errors.New("No more events")
				}

				if end > len(testdata.Events) {
					end = len(testdata.Events)
				}

				result := testdata.Events[offset:end]
				offset = end

				return result, nil
			})

		eventRepo.
			EXPECT().
			Remove(gomock.AssignableToTypeOf([]uint64{})).
			AnyTimes().
			DoAndReturn(func(idList []uint64) error {
				for _, id := range idList {
					removed <- streaming.LikeEvent{ID: id}
				}

				return nil
			})
	})

	sender := testFixture.EventSenderMock(func(sender *mocks.MockEventSender) {
		sender.
			EXPECT().
			Send(gomock.AssignableToTypeOf(&streaming.LikeEvent{})).
			Times(len(testdata.Events)).
			DoAndReturn(func(event *streaming.LikeEvent) error {
				return nil
			})
	})

	return repo, sender, removed
}

func compare(first []streaming.LikeEvent, second []streaming.LikeEvent) bool {
	sort.SliceStable(first, func(i, j int) bool {
		return first[i].ID < first[j].ID
	})
	sort.SliceStable(second, func(i, j int) bool {
		return second[i].ID < second[j].ID
	})

	if len(first) != len(second) {
		return false
	}

	for i := 0; i < len(first); i++ {
		if first[i] != second[i] {
			return false
		}
	}

	return true
}

func TestStart(t *testing.T) {
	repo, sender, removedChan := setup(t)

	cfg := retranslator2.Config{
		ChannelSize:    512,
		ConsumerCount:  2,
		ConsumeSize:    10,
		ConsumeTimeout: 100 * time.Millisecond,
		ProducerCount:  2,
		WorkerCount:    2,
		Repo:           repo,
		Sender:         sender,
	}

	retranslator := retranslator2.NewRetranslator(cfg)

	retranslator.Start(context.Background())
	time.Sleep(time.Second)
	<-retranslator.Stopped()

	close(removedChan)
	result := make([]streaming.LikeEvent, 0, len(testdata.Events))
	for event := range removedChan {
		result = append(result, event)
	}

	if !compare(testdata.Events, result) {
		t.Fail()
	}
}
