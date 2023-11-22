package superapp

type WorkerPool struct {
	workers chan struct{}
}

func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		workers: make(chan struct{}, maxWorkers),
	}
}

func (wp *WorkerPool) Execute(task func()) {
	wp.workers <- struct{}{}
	go func() {
		defer func() {
			<-wp.workers
		}()
		task()
	}()
}
