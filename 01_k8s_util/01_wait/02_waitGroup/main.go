package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

func main(){
	f1:= func(ctx context.Context) {
		for {
			select {
			case <- ctx.Done():
				return
			default:
				fmt.Println("hi11")
				time.Sleep(time.Second)
			}
		}
	}
	wg := wait.Group{}
	ctx, cancel := context.WithCancel(context.Background())
	wg.StartWithContext(ctx,f1)
	time.Sleep(time.Second*3)
	cancel()
	wg.Wait()
}
