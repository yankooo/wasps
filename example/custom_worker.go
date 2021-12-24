package main

import (
	"context"
	"github.com/yankooo/wasps"
	"log"
	"time"
)

type customWorker struct {
	job   chan *wasps.Job
	close chan struct{}
}

type customTypeFn func() error

func (c *customWorker) Do(job *wasps.Job) {
	f := job.PayLoad.(customTypeFn)
	_ = f()
}

func (c *customWorker) JobChan() chan *wasps.Job {
	return c.job
}

func (c *customWorker) StopChan() chan struct{} {
	return c.close
}

func customWorkerExample() {
	pool := wasps.New(5, func() wasps.Worker {
		return &customWorker{
			job:   make(chan *wasps.Job),
			close: make(chan struct{}),
		}
	})
	defer pool.Release()

	ctx, _ := context.WithTimeout(context.TODO(), time.Second*2)
	var task customTypeFn = func() error {
		panic("custome")
		return nil
	}

	pool.SubmitWithContext(ctx, task, wasps.WithRecoverFn(func(r interface{}) {
		log.Printf("catch panic : %+v", r)
	}))

	time.Sleep(time.Second * 2)
}
