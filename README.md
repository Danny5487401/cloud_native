# kubernetes_learning(K8S阿里学习笔记)

![](img/.01_basis_idea/k8s_roadMap.png)

## [第一章 k8s架构及基本概念](chapter01_k8s_basic/01_kube_structure_n_basic_idea.md)

## [第二章 Pod基本单元及相关使用](chapter01_k8s_basic/02_pod.md)

## [第三章 应用编排基本概念](chapter01_k8s_basic/03_resource_object.md)

## [第四章 应用编排deployment](chapter01_k8s_basic/04_deployment.md)

## [第五章 应用编排Job&CronJobs和DaemonSet](chapter01_k8s_basic/05_Job_n_daemonSet.md)

## [第六章 ConfigMap](chapter01_k8s_basic/06_configMap.md)

## [第七章 应用存储和数据卷](chapter01_k8s_basic/07_volume.md)

## [第八章 k8s网络](chapter01_k8s_basic/08_k8s_network_model.md)

## [第九章 Service](chapter01_k8s_basic/09_service.md)

## [第十章 深入linux容器](chapter01_k8s_basic/10_container.md)

## [第十一章 容器运行时接口Container runtime interface](chapter01_k8s_basic/11_cri.md)

## [第十二章 scaler自动弹性伸缩](chapter01_k8s_basic/12_scaler.md)

## [第十三章 kubelet](chapter01_k8s_basic/13_kubelet.md)

## [第十四章 informer机制](chapter01_k8s_basic/14_informer.md)

## 源码篇
- [1 wait工具包](chapter04_k8s_pkg/01_k8s_util/01_wait/wait_util.md)
  - [wait.Until使用](chapter04_k8s_pkg/01_k8s_util/01_wait/01_util/main.go)
  - [wait.Group{}](chapter04_k8s_pkg/01_k8s_util/01_wait/02_waitGroup/main.go)
- 2 sets工具包
  - [判断两个map的key是否重合](chapter04_k8s_pkg/01_k8s_util/02_sets/main.go)
- [3 k8s使用的web框架：go-restful 源码分析](chapter04_k8s_pkg/02_k8s_restful/go-restful.md)
- [4 client-go中rest模块源码分析](chapter04_k8s_pkg/01_k8s_util/03_restclient/rest.md)
  - [4.1 使用restclient与k8s交互](chapter04_k8s_pkg/01_k8s_util/03_restclient/main.go)
