# k8s.io/client-go 中rest模块源码分析
对于不同的kubernetes版本使用标签 v0.x.y 来表示对应的客户端版本。

client-go的客户端对象有4个
- RESTClient：是最基础的基础架构，其作用是将是使用了http包进行封装成RESTClient。位于rest 目录，RESTClient封装了资源URL的通用格式，例如Get()、Put()、Post() Delete()。是与Kubernetes API的访问行为提供的基于RESTful方法进行交互基础架构。
    - 同时支持Json 与 protobuf
    - 支持所有的原生资源和CRD
- ClientSet: Clientset基于RestClient进行封装对 Resource 与 version 管理集合，默认情况下，不能操作CRD资源，但是通过client-gen代码生成的话，也是可以操作CRD资源的。
- DynamicClient:不仅能对K8S内置资源进行处理，还可以对CRD资源进行处理，不需要client-gen生成代码即可实现。
- DiscoveryClient：用于发现kube-apiserver所支持的资源组、资源版本、资源信息（即Group、Version、Resources）


## client-go 目录介绍

client-go的每一个目录都是一个go package

- kubernetes 包含与Kubernetes API所通信的客户端集
- discovery 用于发现kube-apiserver所支持的api
- dynamic 包含了一个动态客户端，该客户端能够对kube-apiserver任意的API进行操作。
- transport 提供了用于设置认证和启动链接的功能
- tools/cache: 一些 low-level controller与一些数据结构如fifo，reflector等



