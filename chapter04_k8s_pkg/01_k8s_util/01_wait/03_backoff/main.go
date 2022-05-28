package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)


func main(){
	var DefaultRetry = wait.Backoff{
		Steps:    5,
		Duration: 1 * time.Second,
		Factor:   0,
		Jitter:   0,
	}

	fmt.Println(wait.ExponentialBackoff(DefaultRetry,func() (bool, error){
		fmt.Println(time.Now())
		return false,nil
	}))
}
