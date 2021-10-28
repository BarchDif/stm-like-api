package fixture

import (
	"github.com/BarchDif/stm-like-api/internal/app/consumer"
	"github.com/BarchDif/stm-like-api/internal/app/producer"
	"github.com/BarchDif/stm-like-api/internal/app/repo"
	"github.com/BarchDif/stm-like-api/internal/app/sender"
	"github.com/BarchDif/stm-like-api/internal/mocks"
	streaming "github.com/BarchDif/stm-like-api/internal/model"
	"github.com/golang/mock/gomock"
	"math/rand"
	"sync"
	"testing"
)

type TestFixture interface {
	ConsumerMock(setup func(*mocks.MockConsumer)) consumer.Consumer
	ProducerMock(func(*mocks.MockProducer)) producer.Producer
	LikeRepoMock(*sync.Mutex, map[uint64]int) repo.EventRepo
	EventSenderMock(func(*mocks.MockEventSender)) sender.EventSender
}

type fixture struct {
	controller *gomock.Controller
}

func NewFixture(t *testing.T) TestFixture {
	controller := gomock.NewController(t)

	return &fixture{
		controller: controller,
	}
}

func (f fixture) ConsumerMock(setup func(*mocks.MockConsumer)) consumer.Consumer {
	consumer := mocks.NewMockConsumer(f.controller)

	if setup != nil {
		setup(consumer)
	}

	return consumer
}

func (f fixture) ProducerMock(setup func(*mocks.MockProducer)) producer.Producer {
	producer := mocks.NewMockProducer(f.controller)

	if setup != nil {
		setup(producer)
	}

	return producer
}

func (f fixture) LikeRepoMock(mu *sync.Mutex, match map[uint64]int) repo.EventRepo {
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

func (f fixture) EventSenderMock(setup func(*mocks.MockEventSender)) sender.EventSender {
	sender := mocks.NewMockEventSender(f.controller)

	if setup != nil {
		setup(sender)
	}

	return sender
}
