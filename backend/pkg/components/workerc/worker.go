package workerc

import (
	"inventory-movement-processing/pkg/logger"
	sctx "inventory-movement-processing/pkg/service_context"
	"runtime"
	"sync"
)

type Job func()

type workerPool struct {
	id         string
	logger     logger.Logger
	numWorkers int
	queueSize  int
	jobQueue   chan Job
	wg         sync.WaitGroup
	quit       chan struct{}
}

func NewPool(id string, numWorkers int, queueSize int) *workerPool {
	return &workerPool{
		id:         id,
		numWorkers: numWorkers,
		queueSize:  queueSize,
	}
}

// component interface

func (wp *workerPool) ID() string { return wp.id }

func (wp *workerPool) InitFlags() {
	if wp.numWorkers <= 0 {
		wp.numWorkers = runtime.NumCPU() * 4
	}

	if wp.queueSize <= 0 {
		wp.queueSize = wp.numWorkers * 10
	}
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

		go func(workerID int) {
			defer wp.wg.Done()

			for {
				select {

				case job, ok := <-wp.jobQueue:
					if !ok {
						return
					}

					func() {
						defer func() {
							if r := recover(); r != nil {
								if wp.logger != nil {
									wp.logger.Error("worker panic", r)
								}
							}
						}()

						job()
					}()

				case <-wp.quit:
					return
				}
			}
		}(i)
	}
}

func (wp *workerPool) Submit(job Job) {
	wp.jobQueue <- job
}
