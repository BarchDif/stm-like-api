package workerpool

type WorkerPool struct {
}

func New(cnt int) *WorkerPool {
	return &WorkerPool{}
}

func (pool *WorkerPool) Submit(f func()) {

}

func (pool *WorkerPool) StopWait() {

}
