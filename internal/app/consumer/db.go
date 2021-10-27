package consumer

//go:generate mockgen -destination=../../mocks/consumer_mock.go -package=mocks github.com/BarchDif/stm-like-api/internal/app/consumer Consumer

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/app/repo"
	"github.com/BarchDif/stm-like-api/internal/model"
	"sync"
	"time"
)

type Consumer interface {
	Start(ctx context.Context)
	Cancel() <-chan bool
}

type consumer struct {
	n      uint64
	events chan<- streaming.LikeEvent

	repo repo.EventRepo

	batchSize uint64
	timeout   time.Duration

	done      chan bool
	cancel    func()
	cancelled chan bool
	wg        *sync.WaitGroup
}

type Config struct {
	n         uint64
	events    chan<- streaming.LikeEvent
	repo      repo.EventRepo
	batchSize uint64
	timeout   time.Duration
}

func NewDbConsumer(
	n uint64,
	batchSize uint64,
	consumeTimeout time.Duration,
	repo repo.EventRepo,
	events chan<- streaming.LikeEvent) Consumer {

	wg := &sync.WaitGroup{}
	done := make(chan bool)
	stopped := make(chan bool)

	return &consumer{
		n:         n,
		batchSize: batchSize,
		timeout:   consumeTimeout,
		repo:      repo,
		events:    events,
		wg:        wg,
		done:      done,
		cancelled: stopped,
	}
}

func (c *consumer) Start(ctx context.Context) {
	childContext, stopFunc := context.WithCancel(ctx)
	c.cancel = stopFunc

	for i := uint64(0); i < c.n; i++ {
		c.wg.Add(1)

		go func() {
			defer c.wg.Done()
			ticker := time.NewTicker(c.timeout)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					events, err := c.repo.Lock(c.batchSize)
					if err != nil {
						continue
					}
					for _, event := range events {
						c.events <- event
					}
				case <-childContext.Done():
					return
				}
			}
		}()
	}

	go c.waitCancellation()
}

func (c *consumer) Cancel() <-chan bool {
	c.cancel()

	return c.cancelled
}

func (c *consumer) waitCancellation() {
	c.wg.Wait()

	close(c.events)

	c.cancelled <- true
}
