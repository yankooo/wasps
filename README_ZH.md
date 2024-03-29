# wasps

[![Build Status](https://travis-ci.com/yankooo/wasps.svg?branch=master)](https://travis-ci.com/yankooo/wasps) [![Go Report Card](https://goreportcard.com/badge/github.com/yankooo/wasps)](https://goreportcard.com/report/github.com/yankooo/wasps) [![Codecov](https://img.shields.io/codecov/c/github/yankooo/wasps/master)](https://codecov.io/gh/yankooo/wasps) [![Doc for wasps](https://img.shields.io/badge/go.dev-doc-007d9c?style=flat&logo=appveyor)](https://pkg.go.dev/github.com/yankooo/wasps?tab=doc) [![GitHub](https://img.shields.io/github/license/yankooo/wasps)](https://github.com/yankooo/wasps/blob/master/LICENSE)

[英文](README.md) | 中文

## 简介

`wasps`是一个轻量级的 goroutine 池，实现了对多个 goroutine 的调度管理。

## 功能

- 自动调度goroutine。
- 提供了常用的接口：任务提交、获取运行中的 goroutine 数量、动态调整 Pool 大小、释放 Pool。
- 提供了回调类型的协程池、串行化工作的协程池、自定义工作协程的协程池。
- 支持自定义工作协程，支持任务协程的panic处理以及闭包函数的自定义传参。
- 异步机制

## 文档

https://godoc.org/github.com/yankooo/wasps

## 安装

``` go
go get github.com/yankooo/wasps
```

## 使用
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
    // 初始化协程池
    pool := wasps.NewCallback(5)
    defer func() {
        pool.Release()
    }()

    var num = 10
    ctx, _ := context.WithTimeout(context.TODO(), 1*time.Second)

    var wg sync.WaitGroup
    wg.Add(3)
    
    // 提交第一个任务
    var f1 wasps.DefaultWorkerPayLoad = func(ctx context.Context, args ...interface{}) {
        defer wg.Done()
        num := args[0].(int)
        log.Printf("first submit %d", num*2)
    }
    pool.SubmitWithContext(ctx, f1, wasps.WithArgs(num))
    
    // 提交第二个任务
    var f2 wasps.DefaultWorkerPayLoad = func(ctx context.Context, args ...interface{}) {
        defer wg.Done()
        num := args[0].(int)
        log.Printf("second submit %d", num)
    }
    pool.Submit(f2, wasps.WithArgs(num), wasps.WithRecoverFn(func(r interface{}) { fmt.Printf("catch panic: %+v\n", r) }))
    
    // 提交第三个任务
    num = 200
    var f3 wasps.DefaultWorkerPayLoad = func(ctx context.Context, args ...interface{}) {
        defer wg.Done()
        log.Printf("third submit %d", num)
    }
    pool.Submit(f3)

    wg.Wait()

    time.Sleep(time.Second*3)
    
    // 直接提交单个任务，不用sync group
    pool.Do(func(ctx context.Context, args ...interface{}) {
        log.Printf("four submit %+v", args)
    })

    time.Sleep(time.Hour)
}
```