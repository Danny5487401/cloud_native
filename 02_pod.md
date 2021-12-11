# Pod
![](img/.02_pod_images/Pod.png)
## 为什么Pod必须是原子调度单位
![](img/.02_pod_images/pod_as_atomic_unit.png)
![](img/.02_pod_images/pod_relation.png)

## Pod需要解决的问题
如何让一个pod里面的多个容器之间高效的共享资源和数据?
但是容器之间原本是被linux Namespace 和cgroups隔离开的
### 解决
1. 共享网络
![](img/.02_pod_images/share_network.png)
2. 共享存储
![](img/.02_pod_images/share_storage.png)

## 容器设计模式

### SideCar
![](img/.02_pod_images/sideCar.png)
应用    
![](img/.02_pod_images/sideCar_application.png)
![](img/.02_pod_images/sideCar_proxy_container.png)
![](img/.02_pod_images/sideCar_adopter_container.png)