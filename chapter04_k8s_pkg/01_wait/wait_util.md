<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [kubernetes 工具函数 wait 使用](#kubernetes-%E5%B7%A5%E5%85%B7%E5%87%BD%E6%95%B0-wait-%E4%BD%BF%E7%94%A8)
  - [分类](#%E5%88%86%E7%B1%BB)
    - [until](#until)
    - [poll](#poll)
    - [第三方应用 WaitForCacheSync -->argo workflow](#%E7%AC%AC%E4%B8%89%E6%96%B9%E5%BA%94%E7%94%A8-waitforcachesync---argo-workflow)
  - [backoff 退让](#backoff-%E9%80%80%E8%AE%A9)
  - [wait.Group 源码](#waitgroup-%E6%BA%90%E7%A0%81)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# kubernetes 工具函数 wait 使用


## 分类
wait库内的各种function，大体来说都是以轮询的形式，根据时间间隔、条件判断，来确定工具执行函数是否应被继续执行

|    条件类型    |                                               说明                                                |
|:----------:|:-----------------------------------------------------------------------------------------------:| 
|   Until类   |                    用得最多的类型，一般以一条chan struct{} 或context Done接收done信号作为终止轮询的依据                    | 
|   poll类    |                             两条channel，一条用作传递单次执行信号用来轮询，一条用作传递done信号                             | 
|  Backoff类  |                      每间隔一定的时长执行一次回溯函数，一般情况下，间隔时长随着回溯次数递增而倍数级延长，但间隔时长也会有上限值                      |







### until 
```go
// k8s.io/apimachinery@v0.26.15/pkg/util/wait/wait.go

// Until函数每period会调度f函数，如果stopCh中有停止信号，则退出。
// 当程序运行时间超过period时，也不会退出调度循环，该特性和Ticker相同。底层使用Timer实现。
func Until(f func(), period time.Duration, stopCh <-chan struct{}) {
    JitterUntil(f, period, 0.0, true, stopCh)
}


func NonSlidingUntil(f func(), period time.Duration, stopCh <-chan struct{}) {
	JitterUntil(f, period, 0.0, false, stopCh)
}
```

Until和NonSlidingUntil为一对，UntilWithContext和NonSlidingUntilWithContext为一对，区别只是定时器启动时间点不同，可以简单用下图表示：


最终调用的都是 JitterUntil
```go
func JitterUntil(f func(), period time.Duration, jitterFactor float64, sliding bool, stopCh <-chan struct{}) {
	BackoffUntil(f, NewJitteredBackoffManager(period, jitterFactor, &clock.RealClock{}), sliding, stopCh)
}
```
> If sliding is true, the period is computed after f runs. If it is false then period includes the runtime for f.
> sliding为true时，时间不包括f运行时间，即f函数执行完成后计时周期。


```go
// k8s.io/apimachinery@v0.26.15/pkg/util/wait/wait.go

func BackoffUntil(f func(), backoff BackoffManager, sliding bool, stopCh <-chan struct{}) {
	var t clock.Timer
	for {
		select {
		case <-stopCh:
			// f()执行前先判断一次stopCh是否有信号，执行完之后也还要执行一次，说明见下方
			return
		default:
		}

		if !sliding {
			// sliding为false时，时间包括f运行时间
			t = backoff.Backoff()
		}

		func() {
			defer runtime.HandleCrash()
			f()
		}()

		if sliding {
			// sliding为true时，时间不包括f运行时间，即f函数执行完成后计时周期
			t = backoff.Backoff()
		}

		// NOTE: b/c there is no priority selection in golang
		// it is possible for this to race, meaning we could
		// trigger t.C and stopCh, and t.C select falls through.
		// In order to mitigate we re-check stopCh at the beginning
		// of every loop to prevent extra executions of f().
		select {
		case <-stopCh:
			if !t.Stop() {
				<-t.C()
			}
			return
		case <-t.C():
		}
	}
}
```



### poll
```go
// 成功的判断依据
type ConditionFunc func() (done bool, err error)

//  等待 interval 后执行一次 ConditionFunc,之后每隔 interval 时间执行一次 ConditionFunc,直到它返回 true,err
func Poll(interval, timeout time.Duration, condition ConditionFunc) error {
	return PollWithContext(context.Background(), interval, timeout, condition.WithContext())
}



func PollInfinite(interval time.Duration, condition ConditionFunc) error {
	return PollInfiniteWithContext(context.Background(), interval, condition.WithContext())
}

func PollInfiniteWithContext(ctx context.Context, interval time.Duration, condition ConditionWithContextFunc) error {
    return poll(ctx, false, poller(interval, 0), condition)
}
```
```go
func PollUntil(interval time.Duration, condition ConditionFunc, stopCh <-chan struct{}) error {
	ctx, cancel := ContextForChannel(stopCh)
	defer cancel()
	return PollUntilWithContext(ctx, interval, condition.WithContext())
}

func PollUntilWithContext(ctx context.Context, interval time.Duration, condition ConditionWithContextFunc) error {
    return poll(ctx, false, poller(interval, 0), condition)
}
```

```go
// PollImmediate 的作用是设定一个间隔时间和超时时间以及一个函数，在超时时间内没到达一个间隔时间就执行一次函数，直到函数返回 true 或者到达超时时间。
func PollImmediate(interval, timeout time.Duration, condition ConditionFunc) error {
    return PollImmediateWithContext(context.Background(), interval, timeout, condition.WithContext())
}

func PollImmediateWithContext(ctx context.Context, interval, timeout time.Duration, condition ConditionWithContextFunc) error {
	return poll(ctx, true, poller(interval, timeout), condition)
}
```



```go
func poll(ctx context.Context, immediate bool, wait WaitWithContextFunc, condition ConditionWithContextFunc) error {
	if immediate {
		done, err := runConditionWithCrashProtectionWithContext(ctx, condition)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}

	select {
	case <-ctx.Done():
		// returning ctx.Err() will break backward compatibility
		return ErrWaitTimeout
	default:
		return WaitForWithContext(ctx, wait, condition)
	}
}

func WaitForWithContext(ctx context.Context, wait WaitWithContextFunc, fn ConditionWithContextFunc) error {
	waitCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := wait(waitCtx)
	for {
		select {
		case _, open := <-c: // 达到运行条件
			ok, err := runConditionWithCrashProtectionWithContext(ctx, fn)
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			if !open {
				return ErrWaitTimeout
			}
		case <-ctx.Done():
			// returning ctx.Err() will break backward compatibility
			return ErrWaitTimeout
		}
	}
}
```

WaitWithContextFunc 的实现 poller

```go
func poller(interval, timeout time.Duration) WaitWithContextFunc {
	return WaitWithContextFunc(func(ctx context.Context) <-chan struct{} {
        // 执行信号chan
		ch := make(chan struct{})

		go func() {
			defer close(ch)

			tick := time.NewTicker(interval)
			defer tick.Stop()

            // 默认无超时时间设定
			var after <-chan time.Time
			if timeout != 0 {
				// time.After is more convenient, but it
				// potentially leaves timers around much longer
				// than necessary if we exit early.
				timer := time.NewTimer(timeout)
				after = timer.C
				defer timer.Stop()
			}

			for {
				select {
				case <-tick.C:
					// If the consumer isn't ready for this signal drop it and
					// check the other channels.
					select {
					case ch <- struct{}{}:
					default:
					}
				case <-after:
					// 超时
					return
				case <-ctx.Done():
					return
				}
			}
		}()

		return ch
	})
}
```

### 第三方应用 WaitForCacheSync -->argo workflow
```go
// https://github.com/argoproj/argo-workflows/blob/7173a271bb9c59ca67df7a06965eb80afd37c0cb/workflow/controller/controller.go
func (wfc *WorkflowController) createClusterWorkflowTemplateInformer(ctx context.Context) {
	// ...

	if cwftGetAllowed && cwftListAllowed && cwftWatchAllowed {
		wfc.cwftmplInformer = informer.NewTolerantClusterWorkflowTemplateInformer(wfc.dynamicInterface, clusterWorkflowTemplateResyncPeriod)
		go wfc.cwftmplInformer.Informer().Run(ctx.Done())

		// since the above call is asynchronous, make sure we populate our cache before we try to use it later
		if !cache.WaitForCacheSync(
			ctx.Done(),
			wfc.cwftmplInformer.Informer().HasSynced,
		) {
			log.Fatal("Timed out waiting for ClusterWorkflowTemplate cache to sync")
		}
	} else {
		log.Warnf("Controller doesn't have RBAC access for ClusterWorkflowTemplates")
	}
}
```

```go
func WaitForCacheSync(stopCh <-chan struct{}, cacheSyncs ...InformerSynced) bool {
	err := wait.PollImmediateUntil(syncedPollPeriod,
		func() (bool, error) {
			for _, syncFunc := range cacheSyncs {
				if !syncFunc() {
					return false, nil
				}
			}
			return true, nil
		},
		stopCh)
	if err != nil {
		klog.V(2).Infof("stop requested")
		return false
	}

	klog.V(4).Infof("caches populated")
	return true
}
```



## backoff 退让
```go
func ExponentialBackoff(backoff Backoff, condition ConditionFunc) error {
    // 在最外层限定了backoff函数的最多重复执行次数，即等于Steps字段值
	for backoff.Steps > 0 { 
		
		if ok, err := runConditionWithCrashProtection(condition); err != nil || ok {
			// 过程中condition()条件函数执行异常或正常则直接返回
			return err
		}
		if backoff.Steps == 1 {
			break
		}
		time.Sleep(backoff.Step())
	}
	return ErrWaitTimeout
}
```

ExponentialBackoff可以实现在函数执行错误后实现以指数退避方式的延时重试。ExponentialBackoff内部使用的是time.Sleep

```go
type Backoff struct {
    // 表示初始的延时时间
    Duration time.Duration
    // Duration is multiplied by factor each iteration. Must be greater
    // than or equal to zero.
    // 指数退避的因子
    Factor float64

    // 可以看作是偏差因子，该值越大，每次重试的延时的可选区间越大
    Jitter float64
    // The number of steps before duration stops changing. If zero, initial
    // duration is always used. Used for exponential backoff in combination
    // with Factor.
    // 指数退避的步数，可以看作程序的最大重试次数
    Steps int
    // The returned duration will never be greater than cap *before* jitter
    // is applied. The actual maximum cap is `cap * (1.0 + jitter)`.
    // 用于在Factor非0时限制最大延时时间和最大重试次数，为0表示不限制最大延时时间
    Cap time.Duration
}
```


```go
func (b *Backoff) Step() time.Duration {
	if b.Steps < 1 {
		if b.Jitter > 0 {
			return Jitter(b.Duration, b.Jitter)
		}
		return b.Duration
	}
	b.Steps--

	duration := b.Duration

	// calculate the next step
	if b.Factor != 0 {
		b.Duration = time.Duration(float64(b.Duration) * b.Factor)
		if b.Cap > 0 && b.Duration > b.Cap {
			b.Duration = b.Cap
			b.Steps = 0
		}
	}

	if b.Jitter > 0 {
		duration = Jitter(duration, b.Jitter)
	}
	return duration
}
```

```go
func Jitter(duration time.Duration, maxFactor float64) time.Duration {
	if maxFactor <= 0.0 {
		maxFactor = 1.0
	}
	wait := duration + time.Duration(rand.Float64()*maxFactor*float64(duration))
	return wait
}
```

1. Duration= 1 * time.Second, Factor= 0,Jitter=0.5，根据下面计算方式wait := duration + time.Duration(rand.Float64()*maxFactor*float64(duration)),即1+[0.0,1.0)*0.5,预期duration为[1s,1.5s)
2. Duration= 1 * time.Second, Factor= 3,根据下面计算方式b.Duration = time.Duration(float64(b.Duration) * b.Factor)，即duration(1) = duration*3 ，duration(2) = 3 * duration(1)


## wait.Group 源码
创造性地将sync.WaitGroup与chan和ctx结合，实现了协程间同步和等待全部Group中的协程结束的功能。由于StartWithChannel和StartWithContext的入参函数类型比较固定，因此使用上并不通用，但可以作为参考
```go
func (g *Group) Wait() 
func (g *Group) StartWithChannel(stopCh <-chan struct{}, f func(stopCh <-chan struct{}))
func (g *Group) StartWithContext(ctx context.Context, f func(context.Context))
```
