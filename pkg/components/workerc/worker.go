package workerc

import (
	"inventory-movement-processing/pkg/components/loggerc"
	sctx "inventory-movement-processing/pkg/service_context"
	"sync"
)

type Job func()

type workerPool struct {
	id         string
	logger     loggerc.Logger
	numWorkers int
	queueSize  int
	jobQueue   chan Job
	wg         sync.WaitGroup
	quit       chan struct{}
}

func NewPool(id string) *workerPool {
	return &workerPool{
		id: id,
	}
}

// component interface
func (wp *workerPool) ID() string { return wp.id }

func (wp *workerPool) InitFlags() {
	wp.numWorkers = 10
	wp.queueSize = 100
}

func (wp *workerPool) Activate(sc sctx.ServiceContext) error {
	wp.jobQueue = make(chan Job, wp.queueSize)
	wp.quit = make(chan struct{})

	if sc != nil {
		wp.logger = sc.Logger(wp.id)
		wp.logger.Info("init engine...")
	}

	wp.start()
	return nil
}

func (wp *workerPool) Stop() error {
	close(wp.quit)
	wp.wg.Wait()
	return nil
}

// end implemt Component interface in service context

func (wp *workerPool) start() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go func() {
			defer wp.wg.Done()
			for {
				select {
				case job := <-wp.jobQueue:
					job()
				case <-wp.quit:
					return
				}
			}
		}()
	}
}

func (wp *workerPool) Submit(job Job) {
	wp.jobQueue <- job
}
