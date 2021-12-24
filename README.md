# wasps

[![Build Status](https://travis-ci.com/yankooo/wasps.svg?branch=master)](https://travis-ci.com/yankooo/wasps) [![Go Report Card](https://goreportcard.com/badge/github.com/yankooo/wasps)](https://goreportcard.com/report/github.com/yankooo/wasps) [![Codecov](https://img.shields.io/codecov/c/github/yankooo/wasps/master)](https://codecov.io/gh/yankooo/wasps) [![Doc for wasps](https://img.shields.io/badge/go.dev-doc-007d9c?style=flat&logo=appveyor)](https://pkg.go.dev/github.com/yankooo/wasps?tab=doc) [![GitHub](https://img.shields.io/github/license/yankooo/wasps)](https://github.com/yankooo/wasps/blob/master/LICENSE)

English | [中文](README_ZH.md)

## Introduction

`wasps` is a lightweight goroutine pool that implements scheduling management for multiple goroutines.

## Features:

- Automatic scheduling goroutine.
- Provides commonly-used interfaces: task submission, getting the number of running goroutines, dynamically adjusting the size of the pool, and releasing the pool.
- Provide callback type goroutine pool, serialization work goroutine pool, custom work goroutine pool.
- Support custom work goroutine, support panic processing of task goroutine, and custom pass parameter of closure function.
- Asynchronous mechanism.

## Docs

https://godoc.org/github.com/yankooo/wasps

## Installation

``` go
go get github.com/yankooo/wasps
```

## Use
``` go
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
    // init pool
    pool := wasps.NewCallback(5)
    defer func() {
        pool.Release()
    }()

    var num = 10
    ctx, _ := context.WithTimeout(context.TODO(), 1*time.Second)

    var wg sync.WaitGroup
    wg.Add(3)
    
    // submit first task
    var f1 wasps.DefaultWorkerPayLoad = func(ctx context.Context, args ...interface{}) {
        defer wg.Done()
        num := args[0].(int)
        log.Printf("first submit %d", num*2)
    }
    pool.SubmitWithContext(ctx, f1, wasps.WithArgs(num))
    
    // submit secod task
    var f2 wasps.DefaultWorkerPayLoad = func(ctx context.Context, args ...interface{}) {
        defer wg.Done()
        num := args[0].(int)
        log.Printf("second submit %d", num)
    }
    pool.Submit(f2, wasps.WithArgs(num), wasps.WithRecoverFn(func(r interface{}) { fmt.Printf("catch panic: %+v\n", r) }))
    
    // submit thrid task
    num = 200
    var f3 wasps.DefaultWorkerPayLoad = func(ctx context.Context, args ...interface{}) {
        defer wg.Done()
        log.Printf("third submit %d", num)
    }
    pool.Submit(f3)

    wg.Wait()

    time.Sleep(time.Second*3)
    
    // submit single task don't use sync group
    pool.Do(func(ctx context.Context, args ...interface{}) {
        log.Printf("four submit %+v", args)
    })

    time.Sleep(time.Hour)
}
```

