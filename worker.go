package wasps

import (
	"context"
)

// Worker is an interface representing a wasps working agent.
// It will be used to process a job of own job channel, and clean up its resources when being removed from the pool.
type Worker interface {
	// Do will process a job.
	Do(job *Job)

	// return job channel of worker
	JobChan() chan *Job

	// return stop channel of worker
	StopChan() chan struct{}
}

// Job is the struct that represents the smallest unit of worker tasks
type Job struct {
	Ctx       context.Context
	Task      interface{}
	Args      []interface{}
	RecoverFn func(r interface{})
}

// callbackWorker is a minimal Worker implementation that attempts to cast each job into func(...interface{}).
type callbackWorker struct {
	job   chan *Job
	close chan struct{}
}

func (c *callbackWorker) Do(job *Job) {
	select {
	case <-job.Ctx.Done():
		return
	default:
	}

	cb := job.Task.(func(...interface{}))
	cb(job.Args...)
}

func (c *callbackWorker) JobChan() chan *Job {
	return c.job
}

func (c *callbackWorker) StopChan() chan struct{} {
	return c.close
}
