package fixture

import (
	"github.com/BarchDif/stm-like-api/internal/app/consumer"
	"github.com/BarchDif/stm-like-api/internal/app/producer"
	"github.com/BarchDif/stm-like-api/internal/app/repo"
	"github.com/BarchDif/stm-like-api/internal/app/sender"
	"github.com/BarchDif/stm-like-api/internal/app/workerpool"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	streaming "github.com/BarchDif/stm-like-api/internal/model"
	"github.com/golang/mock/gomock"
	"math/rand"
	"sync"
	"testing"
)

type TestFixture struct {
	controller *gomock.Controller
}

func NewFixture(t *testing.T) TestFixture {
	controller := gomock.NewController(t)

	return TestFixture{
		controller: controller,
	}
}

func (f TestFixture) ConsumerMock(setup func(*mocks.MockConsumer)) consumer.Consumer {
	consumer := mocks.NewMockConsumer(f.controller)

	if setup != nil {
		setup(consumer)
	}

	return consumer
}

func (f TestFixture) ProducerMock(setup func(*mocks.MockProducer)) producer.Producer {
	producer := mocks.NewMockProducer(f.controller)

	if setup != nil {
		setup(producer)
	}

	return producer
}

func (f TestFixture) LikeRepoMock(setup func(eventRepo *mocks.MockEventRepo)) repo.EventRepo {
	eventRepo := mocks.NewMockEventRepo(f.controller)

	if setup != nil {
		setup(eventRepo)
	}

	return eventRepo
}

func (f TestFixture) LikeRepoLockMock(mu *sync.Mutex, match map[uint64]int) repo.EventRepo {
	repo := mocks.NewMockEventRepo(f.controller)

	repo.
		EXPECT().
		Lock(gomock.AssignableToTypeOf(uint64(0))).
		AnyTimes().
		DoAndReturn(func(n uint64) ([]streaming.LikeEvent, error) {
			mu.Lock()
			defer mu.Unlock()

			result := make([]streaming.LikeEvent, 0, n)

			for i := uint64(0); i < n; i++ {
				id := rand.Uint64()
				// id should be unique
				if _, ok := match[id]; ok {
					i--
					continue
				}

				result = append(result, streaming.LikeEvent{ID: id})

				match[id] = 0
			}

			return result, nil
		})

	return repo
}

func (f TestFixture) LikeRepoUnlockRemoveMock(unlockResult chan streaming.LikeEvent, removeResult chan streaming.LikeEvent) repo.EventRepo {
	repo := mocks.NewMockEventRepo(f.controller)

	repo.
		EXPECT().
		Unlock(gomock.AssignableToTypeOf([]uint64{})).
		MaxTimes(cap(unlockResult)).
		Do(func(id []uint64) {
			for i := 0; i < len(id); i++ {
				unlockResult <- streaming.LikeEvent{ID: id[i]}
			}
		})

	repo.
		EXPECT().
		Remove(gomock.AssignableToTypeOf([]uint64{})).
		MaxTimes(cap(unlockResult)).
		Do(func(id []uint64) {
			for i := 0; i < len(id); i++ {
				removeResult <- streaming.LikeEvent{ID: id[i]}
			}
		})

	return repo
}

func (f TestFixture) EventSenderMock(setup func(*mocks.MockEventSender)) sender.EventSender {
	sender := mocks.NewMockEventSender(f.controller)

	if setup != nil {
		setup(sender)
	}

	return sender
}

func (f TestFixture) EventSenderSendMock(result chan streaming.LikeEvent) sender.EventSender {
	sender := mocks.NewMockEventSender(f.controller)

	sender.
		EXPECT().
		Send(gomock.AssignableToTypeOf(&streaming.LikeEvent{})).
		Times(cap(result)).
		DoAndReturn(func(event *streaming.LikeEvent) error {
			result <- *event

			return nil
		})

	return sender
}

func (f TestFixture) WorkerPoolSubmitMock() workerpool.WorkerPool {
	pool := mocks.NewMockWorkerPool(f.controller)

	pool.
		EXPECT().
		Submit(gomock.Any()).
		AnyTimes().
		Do(func(f func() error) {
			f()
		})

	return pool
}
