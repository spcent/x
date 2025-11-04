package eventbus

import (
	"container/heap"
	"sync"
)

type asyncWorker struct {
	mu     sync.Mutex
	cond   *sync.Cond
	closed bool
	queue  prioQueue
}

func newAsyncWorker() *asyncWorker {
	aw := &asyncWorker{}
	aw.cond = sync.NewCond(&aw.mu)
	heap.Init(&aw.queue)
	return aw
}

func (aw *asyncWorker) push(j job) {
	aw.mu.Lock()
	if !aw.closed {
		heap.Push(&aw.queue, &prioItem{value: j})
	}

	aw.mu.Unlock()
	aw.cond.Signal()
}

func (aw *asyncWorker) pop() (job, bool) {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	for !aw.closed && aw.queue.Len() == 0 {
		aw.cond.Wait()
	}

	if aw.closed {
		return job{}, false
	}

	it := heap.Pop(&aw.queue).(*prioItem)
	return it.value, true
}

func (aw *asyncWorker) close() {
	aw.mu.Lock()
	aw.closed = true
	aw.mu.Unlock()
	aw.cond.Broadcast()
}
