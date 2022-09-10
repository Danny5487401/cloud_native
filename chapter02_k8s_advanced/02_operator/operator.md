# Operator 


Operator 就可以看成是 CRD 和 Controller 的一种组合特例，Operator 是一种思想，它结合了特定领域知识并通过 CRD 机制扩展了 Kubernetes API 资源，
使用户管理 Kubernetes 的内置资源（Pod、Deployment等）一样创建、配置和管理应用程序，Operator 是一个特定的应用程序的控制器，
通过扩展 Kubernetes API 资源以代表 Kubernetes 用户创建、配置和管理复杂应用程序的实例，通常包含资源模型定义和控制器，通过 Operator 通常是为了实现某种特定软件（通常是有状态服务）的自动化运维。


我们完全可以通过上面的方式编写一个 CRD 对象，然后去手动实现一个对应的 Controller 就可以实现一个 Operator，但是我们也发现从头开始去构建一个 CRD 控制器并不容易，
需要对 Kubernetes 的 API 有深入了解，并且 RBAC 集成、镜像构建、持续集成和部署等都需要很大工作量。为了解决这个问题，社区就推出了对应的简单易用的 Operator 框架，
比较主流的是 k8s sig 小组维护的 kubebuilder 和 Operator Framework(包含CoreOS 开源的 operator-sdk 和 Operator Lifecycle Manager（OLM）)，[但是两者已经融合](https://github.com/kubernetes-sigs/kubebuilder/blob/master/designs/integrating-kubebuilder-and-osdk.md)

## Operator 框架包括

- Operator SDK：使开发人员能够利用其专业知识来构建 Operator，无需了解 Kubernetes API 的复杂性。
- Operator 生命周期管理：监控在 Kubernetes 集群上运行的所有 Operator 的生命周期的安装、更新和管理。
- Operator 计量：为提供专业服务的 Operator 启用使用情况报告。


## Kubebuilder 的工作流程如下：

1. 创建一个新的工程目录
2. 创建一个或多个资源 API CRD 然后将字段添加到资源
3. 在控制器中实现协调循环（reconcile loop），watch 额外的资源
4. 在集群中运行测试（自动安装 CRD 并自动启动控制器）
5. 更新引导集成测试测试新字段和业务逻辑
6. 使用用户提供的 Dockerfile 构建和发布容器


## 参考资料
1. [k8s训练营](https://www.qikqiak.com/k8strain/operator/operator/)
2. [RedHat 对k8s operator 理解](https://www.redhat.com/zh/topics/containers/what-is-a-kubernetes-operator)
3. [sigs 小组Kubebuilder 使用说明](https://book.kubebuilder.io/quick-start.html)