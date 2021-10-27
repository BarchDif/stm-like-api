package producer

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/app/repo"
	"github.com/BarchDif/stm-like-api/internal/app/sender"
	"github.com/BarchDif/stm-like-api/internal/app/workerpool"
	"github.com/BarchDif/stm-like-api/internal/model"
	"math"
	"sync"
	"time"
)

type Producer interface {
	Start(ctx context.Context)
	Cancel() <-chan bool
}

type producer struct {
	n       uint64
	timeout time.Duration

	repo repo.EventRepo

	sender sender.EventSender
	events chan streaming.LikeEvent

	workerPool workerpool.WorkerPool

	formBatch      chan streaming.LikeEvent
	processedBatch []streaming.LikeEvent
	deferredBatch  []streaming.LikeEvent

	senderWaitGroup *sync.WaitGroup
	batcherMutex    *sync.Mutex
	cancel          func()
	cancelled       chan bool
	isCancelled     bool
}

func NewKafkaProducer(
	n uint64,
	repo repo.EventRepo,
	sender sender.EventSender,
	events chan streaming.LikeEvent,
	workerPool workerpool.WorkerPool,
) Producer {
	wg := new(sync.WaitGroup)
	batcherMutex := new(sync.Mutex)
	cancelled := make(chan bool)

	formBatch := make(chan streaming.LikeEvent, n)
	processedBatch := make([]streaming.LikeEvent, 0, n)
	deferredBatch := make([]streaming.LikeEvent, 0, n)

	return &producer{
		n:               n,
		repo:            repo,
		sender:          sender,
		events:          events,
		workerPool:      workerPool,
		formBatch:       formBatch,
		processedBatch:  processedBatch,
		deferredBatch:   deferredBatch,
		senderWaitGroup: wg,
		batcherMutex:    batcherMutex,
		cancelled:       cancelled,
	}
}

func (p *producer) Start(ctx context.Context) {
	childContext, cancelFunc := context.WithCancel(ctx)
	p.cancel = cancelFunc

	for i := uint64(0); i < p.n; i++ {
		p.senderWaitGroup.Add(1)

		go func() {
			defer p.senderWaitGroup.Done()

			for {
				select {
				case <-childContext.Done():
					for event := range p.events {
						p.handleEventSending(&event)
					}

					return
				case event, ok := <-p.events:
					if !ok {
						return
					}

					p.handleEventSending(&event)
				}
			}
		}()
	}

	go p.handleBatches(childContext)

	go p.waitCancellation()
}

func (p *producer) Cancel() <-chan bool {
	p.cancel()

	return p.cancelled
}

func (p *producer) waitCancellation() {
	p.senderWaitGroup.Wait()
	close(p.formBatch)

	p.batcherMutex.Lock()
	defer p.batcherMutex.Unlock()

	p.cancelled <- true
	close(p.cancelled)
}

func (p *producer) handleEventSending(event *streaming.LikeEvent) {
	if err := p.sender.Send(event); err != nil {
		event.Status = streaming.Deferred
	} else {
		event.Status = streaming.Processed
	}

	p.formBatch <- *event
}

func (p *producer) handleBatches(ctx context.Context) {
	p.batcherMutex.Lock()
	defer p.batcherMutex.Unlock()

	for {
		select {
		case <-ctx.Done():
			for event := range p.formBatch {
				p.dispatchEvent(event)
			}

			for len(p.processedBatch) != 0 || len(p.deferredBatch) != 0 {
				if len(p.processedBatch) != 0 {
					p.submitBatch(&p.processedBatch)
				}

				if len(p.deferredBatch) != 0 {
					p.submitBatch(&p.deferredBatch)
				}
			}

			return
		case event, ok := <-p.formBatch:
			if !ok {
				continue
			}

			p.dispatchEvent(event)
		}

		if uint64(len(p.processedBatch)) == p.n {
			p.submitBatch(&p.processedBatch)
		}

		if uint64(len(p.deferredBatch)) == p.n {
			p.submitBatch(&p.deferredBatch)
		}
	}
}

func (p *producer) dispatchEvent(event streaming.LikeEvent) {
	if event.Status == streaming.Processed {
		p.processedBatch = append(p.processedBatch, event)
	} else {
		p.deferredBatch = append(p.deferredBatch, event)
	}
}

func (p *producer) submitBatch(eventBatchPtr *[]streaming.LikeEvent) {
	eventBatch := *eventBatchPtr
	batchSize := int(math.Min(float64(p.n), float64(len(eventBatch))))

	idBatch := make([]uint64, batchSize)
	for i := 0; i < batchSize; i++ {
		idBatch[i] = eventBatch[i].ID
	}
	prevLen := len(eventBatch)
	copy(*eventBatchPtr, eventBatch[batchSize:])
	*eventBatchPtr = eventBatch[:prevLen-batchSize]

	p.workerPool.Submit(func() error {
		err := p.repo.Unlock(idBatch)

		return err
	})
}
