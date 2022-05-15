# kubernetes_learning(K8S阿里学习笔记)

![](img/.01_basis_idea/k8s_roadMap.png)

## [第一章 k8s架构及基本概念](01_kube_structure_n_basic_idea.md)

## [第二章 Pod基本单元及相关使用](02_pod.md)

## [第三章 应用编排基本概念](03_resource_object.md)

## [第四章 应用编排deployment](04_deployment.md)

## [第五章 应用编排Job&CronJobs和DaemonSet](05_Job_n_daemonSet.md)

## [第六章 ConfigMap](06_configMap.md)

## [第七章 应用存储和数据卷](07_volume.md)

## [第八章 k8s网络](08_k8s_network_model.md)

## [第九章 Service](09_service.md)

## [第十章 深入linux容器](10_container.md)

## [第十一章 容器运行时接口Container runtime interface](11_cri.md)

## [第十二章 scaler自动弹性伸缩](12_scaler.md)

## [第十三章 kubelet](13_kubelet.md)

## [第十四章 informer机制](14_informer.md)

## 源码篇
- [1 wait工具包](01_k8s_util/01_wait/wait_util.md)
  - [wait.Until使用](01_k8s_util/01_wait/01_util/main.go)
  - [wait.Group{}](01_k8s_util/01_wait/02_waitGroup/main.go)
- 2 sets工具包
  - [判断两个map的key是否重合](01_k8s_util/02_sets/main.go)
- [3 k8s使用的web框架：go-restful 源码分析](02_k8s_restful/go-restful.md)
- [4 client-go中rest模块源码分析](01_k8s_util/03_restclient/rest.md)
  - [4.1 使用restclient与k8s交互](01_k8s_util/03_restclient/main.go)
- [5 api 与 apimachinery 仓库](05_k8s.io_api/api.md)

