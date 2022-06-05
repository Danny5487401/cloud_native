# CustomResourcesDefinition(crd )

CRD 功能是在 Kubernetes 1.7 版本被引入的，用户可以根据自己的需求添加自定义的 Kubernetes 对象资源。
值得注意的是，这里用户自己添加的 Kubernetes 对象资源都是 native 的、都是一等公民，和 Kubernetes 中自带的、原生的那些 Pod、Deployment 是同样的对象资源。
在 Kubernetes 的 API Server 看来，它们都是存在于 etcd 中的一等资源

同时，自定义资源和原生内置的资源一样，都可以用 kubectl  来去创建、查看，也享有 RBAC、安全功能。用户可以开发自定义控制器来感知或者操作自定义资源的变化

## 案例
### 1. 定义资源
```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: foos.samplecontroller.k8s.io
  # for more information on the below annotation, please see
  # https://github.com/kubernetes/enhancements/blob/master/keps/sig-api-machinery/2337-k8s.io-group-protection/README.md
  annotations:
    "api-approved.kubernetes.io": "unapproved, experimental-only; please get an approval from Kubernetes API reviewers if you're trying to develop a CRD in the *.k8s.io or *.kubernetes.io groups"
spec:
  group: samplecontroller.k8s.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        # schema used for validation
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                deploymentName:
                  type: string
                replicas:
                  type: integer
                  minimum: 1
                  maximum: 10
            status:
              type: object
              properties:
                availableReplicas:
                  type: integer
  names:
    kind: Foo
    plural: foos
  scope: Namespaced

```
- 首先最上面的 apiVersion 就是指 CRD 的一个 apiVersion 声明，声明它是一个 CRD 的需求或者说定义的 Schema
- kind 就是 CustomResourcesDefinition，指 CRD。
- name 是一个用户自定义资源中自己自定义的一个名字。一般我们建议使用“顶级域名.xxx.APIGroup”这样的格式，比如这里就是 foos.samplecontroller.k8s.io。
- spec 用于指定该 CRD 的 group、version
    - names 指的是它的 kind 是什么，比如 Deployment 的 kind 就是 Deployment，Pod 的 kind 就是 Pod，这里的 kind 被定义为了 Foo；
    - plural 字段就是一个昵称，比如当一些字段或者一些资源的名字比较长时，可以用该字段自定义一些昵称来简化它的长度
    - scope 字段表明该 CRD 是否被命名空间管理。比如 ClusterRoleBinding 就是 Cluster 级别的。再比如 Pod、Deployment 可以被创建到不同的命名空间里，那么它们的 scope 就是 Namespaced 的。这里的 CRD 就是 Namespaced 的
    
### 2. 使用实例