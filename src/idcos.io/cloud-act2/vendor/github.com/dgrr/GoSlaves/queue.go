package slaves

import (
	"sync"
	"sync/atomic"
	"time"
)

// Queue allows programmer to stack tasks.
type Queue struct {
	closed  bool
	stop    bool
	jobs    []interface{}
	slaves  []*slave
	max     int
	current int32
	locker  sync.RWMutex
	ready   sync.Mutex
	cond    *sync.Cond
	rcond   *sync.Cond
	ok      chan struct{}
}

// max is the maximum goroutines to execute.
// 0 is the same as no limit
func DoQueue(max int, work func(obj interface{})) *Queue {
	queue := &Queue{
		closed: false,
		max:    max,
		ok:     make(chan struct{}),
		jobs:   make([]interface{}, 0),
		slaves: make([]*slave, 0),
	}
	queue.cond = sync.NewCond(&queue.locker)
	queue.rcond = sync.NewCond(&queue.ready)

	go hardClean(queue.locker, &queue.slaves)

	go func() {
		// selected slave
		var s *slave
		var c interface{}
		m := int32(max)
		var i, l int
		for {
			queue.locker.Lock()
			l = len(queue.jobs)
			if l == 0 {
				if queue.closed {
					queue.ok <- struct{}{}
					return
				}
				queue.cond.Wait()
			}
			for i, l = 0, len(queue.jobs); i < l; i++ {
				if l == 0 && queue.closed {
					queue.ok <- struct{}{}
					return
				}
				// getting job to do
				c = queue.jobs[0]
				queue.ready.Lock()
				if len(queue.slaves) == 0 {
					if atomic.LoadInt32(&queue.current) >= m {
						queue.rcond.Wait()
					} else {
						s = pool.Get().(*slave)
						atomic.AddInt32(&queue.current, 1)
						go func(sv *slave) {
							defer atomic.AddInt32(&queue.current, -1)
							defer pool.Put(sv)
							var w interface{}
							for w = range sv.ch {
								if w == nil {
									return
								}
								work(w)
								sv.lastUsage = time.Now()
								queue.ready.Lock()
								queue.slaves = append(queue.slaves, sv)
								queue.ready.Unlock()
								queue.rcond.Signal()
							}
						}(s)
					}
				}
				queue.ready.Unlock()
				if s == nil {
					s = queue.slaves[0]
					queue.slaves = append(queue.slaves[:0], queue.slaves[1:]...)
				}
				queue.jobs = append(queue.jobs[:0], queue.jobs[1:]...)
				// parsing job
				queue.locker.Unlock()
				s.ch <- c
				queue.locker.Lock()
				s = nil
			}
			queue.locker.Unlock()
		}
	}()

	return queue
}

func (queue *Queue) Serve(job interface{}) {
	queue.locker.Lock()
	queue.jobs = append(queue.jobs, job)
	if queue.stop {
		queue.locker.Unlock()
		return
	}
	queue.locker.Unlock()
	queue.cond.Signal()
}

// Stop stops goroutine execution.
// When job is sended to pool it will be stored.
// Jobs will be processed where Resume() is called.
func (queue *Queue) Stop() {
	queue.locker.Lock()
	queue.stop = true
	queue.locker.Unlock()
}

// Resume continues goroutine work.
func (queue *Queue) Resume() {
	queue.locker.Lock()
	queue.stop = false
	queue.locker.Unlock()
}

func (queue *Queue) Close() {
	queue.closed = true
	queue.cond.Signal()
	<-queue.ok
}
