package wasps

import (
	"context"
	"sync"
)

// Pool is a struct that manages a collection of workers, each with their own
// goroutine. The Pool can initialize, expand, compress and close the workers,
// as well as asynchronously process submitted tasks.
type Pool struct {
	capacity   int
	taskChan   chan *Job
	ctor       func() Worker
	workerChan chan Worker
	stop       chan struct{ size int }
	stat       chan *Stats
	m          sync.RWMutex
}

// Capacity returns the current capacity of the pool.
func (p *Pool) Capacity() int {
	return p.getCap()
}

func (p *Pool) getCap() int {
	p.m.RLock()
	c := p.capacity
	p.m.RUnlock()
	return c
}

func (p *Pool) setCap(cap int) {
	p.m.Lock()
	p.capacity = cap
	p.m.Unlock()
}

// New creates a new Pool of workers that starts with a number of workers. You must
// provide a constructor function that creates new Worker types and when you
// change the capacity of the pool the constructor will be called to create a new Worker.
// capacity - how many workers will be created for this pool and size of the pool.
// ctor - constructor function that creates new Worker types
func New(capacity int, ctor func() Worker) *Pool {
	p := &Pool{
		capacity:   capacity,
		ctor:       ctor,
		workerChan: make(chan Worker, capacity),
		taskChan:   make(chan *Job, 1),
		stop:       make(chan struct{ size int }),
		stat:       make(chan *Stats),
	}

	go p.startSchedule()
	for i := 0; i < capacity; i++ {
		p.createWork()
	}

	return p
}

// NewCallback creates a new Pool of workers where workers cast the Job payload
// into a func() and runs it, or returns ErrNotFunc if the cast failed.
func NewCallback(capacity int) *Pool {
	return New(capacity, func() Worker {
		return &callbackWorker{
			job:   make(chan *Job),
			close: make(chan struct{}),
		}
	})
}

// Task payload
type payLoad interface{}

// Submit will submit a task to the task queue of the goroutine pool.
func (p *Pool) Submit(pl payLoad, opts ...TaskOption) {
	p.SubmitWithContext(context.TODO(), pl, opts...)
}

// SubmitWithContext will submit a task to the task queue of the coroutine pool, accompanied by a context.
// Before the task is executed, if the context is canceled, the task will not be executed.
func (p *Pool) SubmitWithContext(ctx context.Context, pl payLoad, opts ...TaskOption) {
	p.submit(ctx, pl, opts...)
}

func (p *Pool) submit(ctx context.Context, task interface{}, opts ...TaskOption) {
	for _, opt := range opts {
		opt.apply(defaultTaskOptions)
	}
	p.taskChan <- &Job{
		Ctx:       ctx,
		Task:      task,
		Args:      defaultTaskOptions.Args,
		RecoverFn: defaultTaskOptions.RecoverFn}
}

// Stats contains running pool Infos.
type Stats struct {
	Cap         int // goroutine pool capacity
	IdleWorker  int // Number of work goroutines in idle state
	WaitingTask int // Number of tasks waiting to be processed
}

// Stats returns information during the running of the goroutine pool.
func (p *Pool) Stats() *Stats {
	if p.isClosed() {
		return &Stats{}
	}
	p.stat <- &Stats{}
	return <-p.stat
}

// SetCapacity changes the capacity of the pool and the total number of workers in the Pool. This can be called
// by any goroutine at any time unless the Pool has been stopped, in which case
// a panic will occur.
func (p *Pool) SetCapacity(size int) {
	if size == p.getCap() {
		return
	}

	if size == 0 {
		p.Release()
	} else if size > p.getCap() {
		for i := size - p.getCap(); i > 0; i-- {
			p.createWork()
		}
	} else {
		p.stop <- struct{ size int }{size: size}
		<-p.stop
	}

	p.setCap(size)
}

// Release will terminate all workers and close the channel of this Pool.
func (p *Pool) Release() {
	if !p.isClosed() {
		return
	}
	p.stop <- struct{ size int }{size: -1}
	<-p.stop
	p.setCap(0)
	close(p.stop)
}

// isClosed returns whether the coroutine pool is stopped.
func (p *Pool) isClosed() bool {
	select {
	case <-p.stop:
		return true
	default:
	}
	return false
}

// startSchedule will schedule worker queue and task queue.
func (p *Pool) startSchedule() {
	// create task queue and work queue
	var taskQ []*Job
	var workerQ []Worker

	for {
		var activeChan chan *Job
		var activeTask *Job

		// When the task queue and work goroutine queue have data at the same time,
		// take out the task and work goroutine to complete the task
		if len(taskQ) > 0 && len(workerQ) > 0 {
			activeChan = workerQ[0].JobChan()
			activeTask = taskQ[0]
		}

		select {
		// append task
		case r := <-p.taskChan:
			taskQ = append(taskQ, r)
		// append idle worker
		case w := <-p.workerChan:
			workerQ = append(workerQ, w)
		// complete the task
		case activeChan <- activeTask:
			taskQ = taskQ[1:]
			workerQ = workerQ[1:]
		// export pool status
		case <-p.stat:
			p.stat <- &Stats{
				Cap:         p.getCap(),
				IdleWorker:  len(workerQ),
				WaitingTask: len(taskQ),
			}
		// reduce one or all worker
		case s := <-p.stop:
			var start int
			var idleWorkerNum = len(workerQ)
			if s.size == -1 {
				start = 0
			} else if s.size == 0 {
				start = idleWorkerNum
			} else {
				start = idleWorkerNum - s.size
			}

			for ; start < idleWorkerNum; start++ {
				worker := workerQ[0]
				worker.StopChan() <- struct{}{}
				<-worker.StopChan()
				workerQ = workerQ[1:]
			}

			p.stop <- struct{ size int }{}
			if s.size == -1 {
				// reducing all workers means releasing the goroutine pool and should stop scheduling
				return
			}
		}
	}
}

// workerReady notify pool that worker is an idle worker.
func (p *Pool) workerReady(w Worker) {
	p.workerChan <- w
}

// createWork to create a new Worker by constructor func.
func (p *Pool) createWork() {
	worker := p.ctor()
	go func() {
		for {
			p.workerReady(worker)
			select {
			case job := <-worker.JobChan():
				func() {
					// This defer function will try to catches a crash
					defer func() {
						if r := recover(); r != nil {
							if job.RecoverFn != nil {
								job.RecoverFn(r)
							}
						}
					}()
					worker.Do(job)
				}()
			case <-worker.StopChan():
				worker.StopChan() <- struct{}{}
				return
			}
		}
	}()
}
