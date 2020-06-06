package wasps

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestSubmitWithContextWithRecover(t *testing.T) {
	pool := NewCallback(1)
	defer func() {
		pool.Release()
		t.Log("pool released")
	}()

	var num = 10
	ctx, _ := context.WithTimeout(context.TODO(), 1*time.Second)

	var wait = make(chan int)

	pool.SubmitWithContext(ctx, func(args ...interface{}) {
		defer func() { wait <- 0 }()
		num := args[0].(int)
		assert.Equal(t, 20, num*2)
		panic("custom panic")
	}, WithArgs(num), WithRecoverFn(func(r interface{}) {
		assert.Equal(t, "custom panic", r.(string))
		t.Logf("catch panic: %+v", r)
	}))
	<-wait
}

func TestParallelJobs(t *testing.T) {
	pool := NewCallback(2)
	defer func() {
		pool.Release()
		t.Log("pool released")
	}()
	var num = 10
	ctx, _ := context.WithTimeout(context.TODO(), 1*time.Second)

	var wg sync.WaitGroup
	wg.Add(3)
	pool.SubmitWithContext(ctx, func(args ...interface{}) {
		defer wg.Done()
		num := args[0].(int)
		t.Log("first submit", num*2)
		assert.Equal(t, 20, num*2)
	}, WithArgs(num))

	pool.Submit(func(args ...interface{}) {
		defer wg.Done()
		num := args[0].(int)
		assert.Equal(t, 10, num)
	}, WithArgs(num), WithRecoverFn(func(r interface{}) { fmt.Printf("catch panic: %+v\n", r) }))

	num = 200
	pool.Submit(func(args ...interface{}) {
		defer wg.Done()
		assert.Equal(t, 200, num)
	})

	wg.Wait()
}

func TestPool_SetSize(t *testing.T) {
	pool := NewCallback(10)
	defer pool.Release()

	time.Sleep(time.Second * 5)
	stats := pool.Stats()
	t.Logf("now pool stat: %+v", stats)

	pool.SetCapacity(5)

	stats = pool.Stats()
	t.Logf("now pool stat: %+v", stats)

	pool.SetCapacity(15)
	time.Sleep(time.Second * 3)
	stats = pool.Stats()
	t.Logf("now pool stat: %+v", stats)
}

func TestPool_Stats(t *testing.T) {
	pool := NewCallback(10)
	defer pool.Release()
	time.Sleep(time.Second * 1)
	stats := pool.Stats()
	t.Logf("now pool stat: %+v", stats)
	stat := &Stats{
		Cap:         10,
		IdleWorker:  10,
		WaitingTask: 0,
	}
	assert.Equal(t, stat, stats)

	for i := 0; i < 15; i++ {
		pool.Submit(func(args ...interface{}) {
			time.Sleep(time.Minute)
		})
	}

	time.Sleep(time.Second * 3)
	stats = pool.Stats()
	t.Logf("now pool stat: %+v", stats)

	assert.Equal(t, &Stats{
		Cap:         10,
		IdleWorker:  0,
		WaitingTask: 5,
	}, stats)
}

func TestPool_Release(t *testing.T) {
	pool := NewCallback(10)
	defer pool.Release()
	defer pool.Release()
}

//------------------------------------------------------------------------------

func BenchmarkNewCallback(b *testing.B) {
	pool := NewCallback(10)
	defer pool.Release()

	b.ResetTimer()

	var wg sync.WaitGroup
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		pool.Submit(func(args ...interface{}) {
			defer wg.Done()
			_ = args[0].(int)
		})
	}
	wg.Wait()
}

func BenchmarkCallbackPool(b *testing.B) {
	pool := NewCallback(BenchAntsSize)
	defer pool.Release()
	var wg sync.WaitGroup

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			pool.Submit(func(args ...interface{}) {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}

const (
	RunTimes           = 1000000
	BenchParam         = 10
	BenchAntsSize      = 200000
	DefaultExpiredTime = 10 * time.Second
)

func demoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}
