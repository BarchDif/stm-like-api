package workerpool

import (
	"context"
	"github.com/google/uuid"
	"sync"
	"time"
)

//go:generate mockgen -destination=../../mocks/pool_mock.go -package=mocks github.com/BarchDif/stm-like-api/internal/app/workerpool WorkerPool

const maxRetryCount = 5

type WorkerPool interface {
	Submit(func() error)
	Start(context.Context)
	StopWait() chan bool
}

type simpleWorkerPool struct {
	wg          *sync.WaitGroup
	workerCount int
	tasks       chan func() error
	retryList   []retryableTask

	cancel    func()
	cancelled chan bool
}

type retryableTask struct {
	id            uuid.UUID
	retryAttempts int
	task          func() error
}

func New(workerCount int) WorkerPool {
	wg := new(sync.WaitGroup)
	tasks := make(chan func() error, workerCount*2)
	cancelled := make(chan bool)
	retryList := make([]retryableTask, 0)

	return &simpleWorkerPool{
		wg:          wg,
		workerCount: workerCount,
		tasks:       tasks,
		cancelled:   cancelled,
		retryList:   retryList,
	}
}

func (pool *simpleWorkerPool) Submit(task func() error) {
	pool.tasks <- task
}

func (pool *simpleWorkerPool) Start(ctx context.Context) {
	childContext, cancelFunc := context.WithCancel(ctx)
	pool.cancel = cancelFunc

	for i := 0; i < pool.workerCount; i++ {
		pool.wg.Add(1)

		go func() {
			defer pool.wg.Done()

			for {
				select {
				case <-childContext.Done():
					for task := range pool.tasks {
						pool.handleTask(task)
					}

					return
				case task, ok := <-pool.tasks:
					if !ok {
						return
					}

					pool.handleTask(task)
				}
			}
		}()
	}

	go func() {
		pool.wg.Add(1)
		defer pool.wg.Done()

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-childContext.Done():
				for _, retry := range pool.retryList {
					pool.handleRetry(&retry)
				}

				return
			case <-ticker.C:
				for _, retry := range pool.retryList {
					pool.handleRetry(&retry)
				}
			}
		}
	}()

	go func() {
		pool.wg.Add(1)
		defer pool.wg.Done()

		<-childContext.Done()
		close(pool.tasks)
	}()

	go func() {
		pool.wg.Wait()

		pool.cancelled <- true
		close(pool.cancelled)
	}()
}

func (pool *simpleWorkerPool) handleTask(task func() error) {
	if err := task(); err != nil {
		pool.scheduleRetry(task)
	}
}

func (pool *simpleWorkerPool) scheduleRetry(task func() error) {
	pool.retryList = append(pool.retryList, retryableTask{
		id:   uuid.New(),
		task: task,
	})
}

func (pool *simpleWorkerPool) handleRetry(retry *retryableTask) {
	retry.retryAttempts++
	err := retry.task()
	if err == nil {
		pool.deleteById(retry.id)

		return
	}

	if retry.retryAttempts > maxRetryCount {
		pool.deleteById(retry.id)
	}
}

func (pool *simpleWorkerPool) deleteById(id uuid.UUID) {
	i := 0
	for ; i < len(pool.retryList); i++ {
		if pool.retryList[i].id == id {
			break
		}
	}

	if i == len(pool.retryList) {
		return
	}

	if i == len(pool.retryList)-1 {
		pool.retryList = pool.retryList[:len(pool.retryList)-1]

		return
	}

	pool.retryList = append(pool.retryList[:i], pool.retryList[i+1:]...)
}

func (pool *simpleWorkerPool) StopWait() chan bool {
	pool.cancel()

	return pool.cancelled
}
