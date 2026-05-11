package workerc

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool_Submit(t *testing.T) {
	pool := NewPool("test-worker-pool", 1, 1)
	pool.InitFlags()

	pool.Activate(nil)
	var count int32

	for i := 0; i < 10; i++ {
		pool.Submit(func() {
			atomic.AddInt32(&count, 1)
		})
	}

	time.Sleep(100 * time.Millisecond)

	pool.Stop()
	if count != 10 {
		t.Errorf("expected 10 jobs done, got %d", count)
	}

}

func TestWorkerPool_Stop(t *testing.T) {
	pool := NewPool("test-pool", 1, 1)
	pool.InitFlags()
	pool.Activate(nil)

	var count int32

	for i := 0; i < 5; i++ {
		pool.Submit(func() {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&count, 1)
		})
	}

	pool.Stop() // phải chờ 5 jobs xong mới return

	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}
