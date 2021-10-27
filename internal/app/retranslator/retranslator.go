package retranslator

import (
	"context"
	"github.com/BarchDif/stm-like-api/internal/app/consumer"
	"github.com/BarchDif/stm-like-api/internal/app/producer"
	"github.com/BarchDif/stm-like-api/internal/app/repo"
	"github.com/BarchDif/stm-like-api/internal/app/sender"
	"github.com/BarchDif/stm-like-api/internal/app/workerpool"
	"github.com/BarchDif/stm-like-api/internal/model"
	"time"
)

type Retranslator interface {
	Start(ctx context.Context)
	Stop()
	Stopped() <-chan interface{}
}

type Config struct {
	ChannelSize uint64

	ConsumerCount  uint64
	ConsumeSize    uint64
	ConsumeTimeout time.Duration

	ProducerCount uint64
	WorkerCount   int

	Repo   repo.EventRepo
	Sender sender.EventSender
}

type retranslator struct {
	events     chan streaming.LikeEvent
	stopped    chan interface{}
	consumer   consumer.Consumer
	producer   producer.Producer
	workerPool workerpool.WorkerPool
}

func NewRetranslator(cfg Config) Retranslator {
	events := make(chan streaming.LikeEvent, cfg.ChannelSize)
	stopped := make(chan interface{}, 1)
	workerPool := workerpool.New(cfg.WorkerCount)

	consumer := consumer.NewDbConsumer(
		cfg.ConsumerCount,
		cfg.ConsumeSize,
		cfg.ConsumeTimeout,
		cfg.Repo,
		events)
	producer := producer.NewKafkaProducer(
		cfg.ProducerCount,
		cfg.Repo,
		cfg.Sender,
		events,
		workerPool)

	return &retranslator{
		events:     events,
		stopped:    stopped,
		consumer:   consumer,
		producer:   producer,
		workerPool: workerPool,
	}
}

func (r *retranslator) Start(ctx context.Context) {
	cancelCtx, _ := context.WithCancel(ctx)

	r.producer.Start(cancelCtx)
	r.consumer.Start(cancelCtx)
	r.workerPool.Start(cancelCtx)

	go func() {
		<-ctx.Done()

		r.Stop()
	}()
}

func (r *retranslator) Stop() {
	<-r.consumer.Cancel()
	<-r.producer.Cancel()
	<-r.workerPool.StopWait()

	r.stopped <- struct{}{}
}

func (r *retranslator) Stopped() <-chan interface{} {
	return r.stopped
}
