# HPA（Horizontal Pod Autoscaler）Pod自动弹性伸缩

K8S通过对Pod中运行的容器各项指标（CPU占用、内存占用、网络请求量）的检测，实现对Pod实例个数的动态新增和减少。

早期的kubernetes版本，只支持CPU指标的检测，因为它是通过kubernetes自带的监控系统heapster实现的。

到了kubernetes 1.8版本后，heapster已经弃用，资源指标主要通过metrics api获取，这时能支持检测的指标就变多了（CPU、内存等核心指标和qps等自定义指标）

Horizontal Pod Autoscaler 实现为一个控制循环，其周期由--horizontal-pod-autoscaler-sync-period选项指定（默认15秒）。

在每个周期内，controller manager都会根据每个HorizontalPodAutoscaler定义的指定的指标去查询资源利用率。 controller manager从资源指标API（针对每个pod资源指标）或自定义指标API（针对所有其他指标）获取指标。

对于每个Pod资源指标（比如：CPU），控制器会从资源指标API中获取相应的指标。然后，如果设置了目标利用率值，则控制器计算利用率值作为容器上等效的资源请求百分比。如果设置了目标原始值，则直接使用原始指标值。然后，控制器将所有目标容器的利用率或原始值（取决于指定的目标类型）取平均值，并产生一个用于缩放所需副本数量的比率。

如果某些Pod的容器未设置相关资源请求，则不会定义Pod的CPU使用率，并且自动缩放器不会对该指标采取任何措施。

## HPA设置

参考配置
```yaml
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: podinfo
spec:
  scaleTargetRef:
    apiVersion: extensions/v1beta1
    kind: Deployment
    name: podinfo
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      targetAverageUtilization: 80
  - type: Resource
    resource:
      name: memory
      targetAverageValue: 200Mi
  - type: Pods
    pods:
      metric:
        name: packets-per-second
      target:
        type: AverageValue
        averageValue: 1k
  - type: Object
    object:
      metric:
        name: requests-per-second
      describedObject:
        apiVersion: networking.k8s.io/v1beta1
        kind: Ingress
        name: main-route
      target:
        type: Value
        value: 10k

```
- minReplicas： 最小pod实例数

- maxReplicas： 最大pod实例数

- metrics： 用于计算所需的Pod副本数量的指标列表

- resource： 核心指标，包含cpu和内存两种（被弹性伸缩的pod对象中容器的requests和limits中定义的指标。）

- object： k8s内置对象的特定指标（需自己实现适配器）

- pods： 应用被弹性伸缩的pod对象的特定指标（例如，每个pod每秒处理的事务数）（需自己实现适配器）

- external： 非k8s内置对象的自定义指标（需自己实现适配器）

## HPA获取自定义指标（Custom Metrics）的底层实现（基于Prometheus）
![](.12_hpa_images/hpa_prometheus.png)

Kubernetes是借助Agrregator APIServer扩展机制来实现Custom Metrics。Custom Metrics APIServer是一个提供查询Metrics指标的API服务（Prometheus的一个适配器），
这个服务启动后，kubernetes会暴露一个叫custom.metrics.k8s.io的API，当请求这个URL时，请求通过Custom Metics APIServer去Prometheus里面去查询对应的指标，然后将查询结果按照特定格式返回。

HPA样例配置：
```yaml
kind: HorizontalPodAutoscaler
apiVersion: autoscaling/v2beta1
metadata:
  name: sample-metrics-app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: sample-metrics-app
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Object
    object:
      target:
        kind: Service
        name: sample-metrics-app
      metricName: http_requests
      targetValue: 100

```

当配置好HPA后，HPA会向Custom Metrics APIServer发送https请求      
```http request
https://<apiserver_ip>/apis/custom-metrics.metrics.k8s.io/v1beta1/namespaces/default/services/sample-metrics-app/http_requests

```
可以从上面的https请求URL路径中得知，这是向 default 这个 namespaces 下的名为 sample-metrics-app 的 service 发送获取 http_requests 这个指标的请求。

Custom Metrics APIServer收到 http_requests 查询请求后，向Prometheus发送查询请求查询 http_requests_total 的值（总请求次数），Custom Metics APIServer再将结果计算成 http_requests （单位时间请求率）返回，实现HPA对性能指标的获取，从而进行弹性伸缩操作。

## 算法
```go
desiredReplicas = ceil[currentReplicas * ( currentMetricValue / desiredMetricValue )]
```
直译为：(当前指标值 ➗ 期望指标值) ✖️ 当前副本数 ，结果再向上取整，最终结果就是期望的副本数量

例如，假设当前指标值是200m ，期望指标值是100m，期望的副本数量就是双倍。因为，200.0 / 100.0 == 2.0

如果当前值是50m，则根据50.0 / 100.0 == 0.5，那么最终的副本数量就是当前副本数量的一半

如果该比率足够接近1.0，则会跳过伸缩

当targetAverageValue或者targetAverageUtilization被指定的时候，currentMetricValue取HorizontalPodAutoscaler伸缩目标中所有Pod的给定指标的平均值。

所有失败的和标记删除的Pod将被丢弃，即不参与指标计算

当基于CPU利用率来进行伸缩时，如果有尚未准备好的Pod（即它仍在初始化），那么该Pod将被放置到一边，即将被保留。

```shell
# 查看autoscalers列表
kubectl get hpa
# 查看具体描述
kubectl describe hpa
# 删除autoscaler
kubectl delete hpa

# 示例：以下命名将会为副本集foo创建一个autoscaler，并设置目标CPU利用率为80%，副本数在2~5之间
kubectl autoscale rs foo --min=2 --max=5 --cpu-percent=80
```