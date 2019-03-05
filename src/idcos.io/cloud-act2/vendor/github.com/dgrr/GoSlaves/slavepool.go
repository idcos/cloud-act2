package slaves

import (
	"sync"
	"time"
)

var (
	defaultTime = time.Second * 10
	pool        = &sync.Pool{
		New: func() interface{} {
			return &slave{
				ch: make(chan interface{}, 1),
			}
		},
	}
)

// This library follows the FILO scheme
type slave struct {
	ch        chan interface{}
	lastUsage time.Time
}

func hardClean(locker sync.RWMutex, slaves *[]*slave) {
	var tmp []*slave
	var c int    // number of workers to be delete
	var i, l int // iterator and len
	for {
		time.Sleep(defaultTime)
		if len(*slaves) == 0 {
			continue
		}
		t := time.Now()
		locker.Lock()
		for i = 0; i < 0; i++ {
			if t.Sub((*slaves)[i].lastUsage) > defaultTime {
				c++
			}
		}
		tmp = append(tmp[:0], (*slaves)[c:]...)
		*slaves = (*slaves)[:c]
		locker.Unlock()
		for i, l = 0, len(tmp); i < l; i++ {
			tmp[i].ch <- nil
		}
	}
}

func softClean(locker sync.RWMutex, slaves *[]*slave) {
	var i, l int // iterator and len
	for {
		time.Sleep(defaultTime)
		locker.Lock()
		if len(*slaves) == 0 {
			locker.Unlock()
			continue
		}
		locker.Unlock()
		for i = 0; i < l; i++ {
			(*slaves)[i].ch <- nil
		}
	}
}

type SlavePool struct {
	lock  sync.RWMutex
	ready []*slave
	stop  chan struct{}
	// SlavePool work
	Work func(interface{})
	// Size is the size of the job receiver channel
	Size int
	// Clean parameter specify if programmer wants to
	// clean the goroutine pool
	//
	// False is faster
	Clean   bool
	ch      chan interface{}
	running bool
}

func (sp *SlavePool) Open() {
	if sp.running {
		panic("Pool is running already")
	}
	clean := sp.Clean

	if sp.Size <= 0 {
		sp.Size = 20
	}
	sp.running = true
	sp.ch = make(chan interface{}, sp.Size)
	sp.stop = make(chan struct{}, 1)
	sp.ready = make([]*slave, 0)

	if clean {
		go hardClean(sp.lock, &sp.ready)
	} else {
		go softClean(sp.lock, &sp.ready)
	}

	go func() {
		var n int
		var job interface{}
		sv := &slave{}
		for {
			select {
			case job = <-sp.ch:
				sp.lock.Lock()
				n = len(sp.ready) - 1
				if n == -1 {
					sv = pool.Get().(*slave)
					go func(s *slave) {
						var job interface{}
						for job = range s.ch {
							if job == nil {
								pool.Put(s)
								return
							}

							sp.Work(job)

							sp.lock.Lock()
							if clean {
								s.lastUsage = time.Now()
							}
							sp.ready = append(sp.ready, s)
							sp.lock.Unlock()
						}
					}(sv)
				} else {
					sv = sp.ready[n]
					sp.ready = sp.ready[:n]
				}
				sp.lock.Unlock()
				sv.ch <- job
			case <-sp.stop:
				close(sp.stop)
				close(sp.ch)
				return
			}
		}
	}()
}

func (sp *SlavePool) Serve(job interface{}) {
	if sp.running {
		sp.ch <- job
	}
}

func (sp *SlavePool) Close() {
	sp.stop <- struct{}{}

	sp.lock.Lock()
	sp.running = false
	sp.lock.Unlock()
}
