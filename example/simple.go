package main

import (
	"context"
	"fmt"
	"github.com/yankooo/wasps"
	"log"
	"sync"
	"time"
)

func main() {
	pool := wasps.NewCallback(5)
	defer func() {
		pool.Release()
	}()

	var num = 10
	ctx, _ := context.WithTimeout(context.TODO(), 1*time.Second)

	var wg sync.WaitGroup

	wg.Add(3)
	pool.SubmitWithContext(ctx, func(args ...interface{}) {
		defer wg.Done()
		num := args[0].(int)
		log.Printf("first submit %d", num*2)
	}, wasps.WithArgs(num))

	pool.Submit(func(args ...interface{}) {
		defer wg.Done()
		num := args[0].(int)
		log.Printf("second submit %d", num)
	}, wasps.WithArgs(num), wasps.WithRecoverFn(func(r interface{}) { fmt.Printf("catch panic: %+v\n", r) }))

	num = 200
	pool.Submit(func(args ...interface{}) {
		defer wg.Done()
		log.Printf("third submit %d", num)
	})

	wg.Wait()

	time.Sleep(time.Second*3)

	num = 200
	pool.Submit(func(args ...interface{}) {
		log.Printf("four submit %+v", args)
	})

	time.Sleep(time.Hour)
}
