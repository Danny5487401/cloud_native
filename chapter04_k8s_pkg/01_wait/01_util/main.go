package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	// 1. 永远运行:我们需要周期性地执行一些动作，比如发送心跳请求给master，那么可以使用 wait 库中的 Forever 功能
	// runForever()

	// 2. 带StopSignal的周期性执行函数
	var counter = 1
	wait.Until(func() {
		if counter > 10 {
			close(stopSignal)
		}
		fmt.Println(time.Now().String())
		counter++
	}, time.Second, stopSignal)

}

var stopSignal = make(chan struct{})

func runForever() {
	wait.Forever(func() {
		fmt.Println(time.Now().String())
	}, time.Second)

}
