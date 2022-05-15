# kubernetes API 的源码实现

kuberentes 组织下有很多个仓库，包括 kubernetes、client-go、api、apimachinery 等. 

- kubernetes 仓库应该是 kubernetes 项目的核心仓库，它包含 kubernetes 控制平面核心组件的源码
- client-go 从名字也不难看出是操作 kubernetes API 的 go 语言客户端；
- api 与 apimachinery 应该是与 kubernetes API 相关的仓库

这里主要关注api 与 apimachinery 这两个仓库。


## api

我们知道 kubernetes 官方提供了多种多样的的 API 资源类型，它们被定义在 k8s.io/api 这个仓库中，作为 kubernetes API 定义的规范地址。

最开始这个仓库只是 kubernetes 核心仓库的一部分，后来 kubernetes API 定义规范被越来越多的其他仓库使用，例如 k8s.io/client-go、k8s.io/apimachinery、k8s.io/apiserver 等，为了避免交叉依赖，所以才把 api 拿出来作为单独的仓库。

k8s.io/api 仓库是只读仓库，所有代码都同步自 https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api 核心仓库。


在 k8s.io/api 仓库定义的 kubernetes API 规范中，Pod 作为最基础的资源类型，一个典型的 YAML 形式的序列化 pod 对象如下所示：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: webserver
  labels:
    app: webserver
spec:
  containers:
  - name: webserver
    image: nginx
    ports:
    - containerPort: 80
```

序列化的 pod 对象最终会被发送到 API-Server 并解码为 Pod 类型的 Go 结构体，同时 YAML 中的各个字段会被赋值给该 Go 结构体。那么，Pod 类型在 Go 语言结构体中是怎么定义的呢？
```go
// /Users/python/go/pkg/mod/k8s.io/api@v0.24.0/core/v1/types.go
type Pod struct {
    // 从TypeMeta字段名可以看出该字段定义Pod类型的元信息，类似于面向对象编程里面
    // Class本身的元信息，类似于Pod类型的API分组、API版本等
    metav1.TypeMeta `json:",inline"`
    // ObjectMeta字段定义单个Pod对象的元信息。每个kubernetes资源对象都有自己的元信息，
    // 例如名字、命名空间、标签、注释等等，kuberentes把这些公共的属性提取出来就是
    // metav1.ObjectMeta，成为了API对象类型的父类
    metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
    // PodSpec表示Pod类型的对象定义规范，最为代表性的就是CPU、内存的资源使用。
    // 这个字段和YAML中spec字段对应
    Spec PodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
    // PodStatus表示Pod的状态，比如是运行还是挂起、Pod的IP等等。Kubernetes会根据pod在
    // 集群中的实际状态来更新PodStatus字段
    Status PodStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}
```
Note: 这里的 metav1 是包 k8s.io/apimachinery/pkg/apis/meta/v1 的别名，本文其他部分的将用 metav1 指代。

- metav1.TypeMeta :前者用于定义资源类型的属性 
- metav1.ObjectMeta : 后者用于定义资源对象的公共属性；
- Spec 用于定义 API 资源类型的私有属性，也是不同 API 资源类型之间的区别所在；
- Status 则是用于描述每个资源对象的状态，这和每个资源类型紧密相关的。
```go
type TypeMeta struct {
    // 包括资源类型的名字
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`
	
	// 以及对应 API 的 schema。这里的 schema 指的是资源类型 API 分组以及版本。
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion"`
}

type ObjectMeta struct {

	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	
	GenerateName string `json:"generateName,omitempty" protobuf:"bytes,2,opt,name=generateName"`


	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`


	SelfLink string `json:"selfLink,omitempty" protobuf:"bytes,4,opt,name=selfLink"`


	UID types.UID `json:"uid,omitempty" protobuf:"bytes,5,opt,name=uid,casttype=k8s.io/kubernetes/pkg/types.UID"`


	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,6,opt,name=resourceVersion"`


	Generation int64 `json:"generation,omitempty" protobuf:"varint,7,opt,name=generation"`
	
	CreationTimestamp Time `json:"creationTimestamp,omitempty" protobuf:"bytes,8,opt,name=creationTimestamp"`


	DeletionTimestamp *Time `json:"deletionTimestamp,omitempty" protobuf:"bytes,9,opt,name=deletionTimestamp"`


	DeletionGracePeriodSeconds *int64 `json:"deletionGracePeriodSeconds,omitempty" protobuf:"varint,10,opt,name=deletionGracePeriodSeconds"`


	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`


	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`


	OwnerReferences []OwnerReference `json:"ownerReferences,omitempty" patchStrategy:"merge" patchMergeKey:"uid" protobuf:"bytes,13,rep,name=ownerReferences"`

	Finalizers []string `json:"finalizers,omitempty" patchStrategy:"merge" protobuf:"bytes,14,rep,name=finalizers"`


	ZZZ_DeprecatedClusterName string `json:"clusterName,omitempty" protobuf:"bytes,15,opt,name=clusterName"`


	ManagedFields []ManagedFieldsEntry `json:"managedFields,omitempty" protobuf:"bytes,17,rep,name=managedFields"`
}
```
metav1.TypeMeta 和 metav1.ObjectMeta 字段从语义上也很好理解，这两个类型作为所有 kubernetes API 资源对象的基类，
每个 API 资源对象需要 metav1.TypeMeta 字段用于描述自己是什么类型， 这样才能构造相应类型的对象，所以相同类型的所有资源对象的 metav1.TypeMeta 字段都是相同的，
但是 metav1.ObjectMeta 则不同，它是定义资源对象实例的属性，即所有资源对象都应该具备的属性。
这部分就是和对象本身相关，和类型无关，所以相同类型的资源对象的 metav1.ObjectMeta 可能是不同的。


在 kubernetes 的 API 资源对象中除了单体对象外，还有对象列表类型，用于描述一组相同类型的对象列表。对象列表的典型应用场景就是列举，对象列表就可以表达一组资源对象

```go
// source code from https://github.com/kubernetes/api/blob/master/core/v1/types.go
type PodList struct {
    // PodList也需要继承metav1.TypeMeta，毕竟对象列表也好、单体对象也好都需要类型属性。
    // PodList比[]Pod类型在yaml或者json表达上多了类型描述，当需要根据YAML构建对象列表的时候，
    // 就可以根据类型描述反序列成为PodList。而[]Pod则不可以，必须确保YAML就是[]Pod序列化的
    // 结果，否则就会报错。这就无法实现一个通用的对象序列化/反序列化。
    metav1.TypeMeta `json:",inline"`
    // 与Pod不同，PodList继承了metav1.ListMeta，metav1.ListMeta是所有资源对象列表类型的父类，
    // ListMeta定义了所有对象列表类型实例的公共属性。
    metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
    // Items字段则是PodList定义的本质，表示Pod资源对象的列表，所以说PodList就是[]Pod基础上加了一些
    // 跟类型和对象列表相关的元信息
    Items []Pod `json:"items" protobuf:"bytes,2,rep,name=items"`
```


### metav1.TypeMeta
```go
// /Users/python/go/pkg/mod/k8s.io/apimachinery@v0.24.0/pkg/runtime/schema/interfaces.go
type ObjectKind interface {
	// SetGroupVersionKind sets or clears the intended serialized kind of an object. Passing kind nil
	// should clear the current setting.
	SetGroupVersionKind(kind GroupVersionKind)
	// GroupVersionKind returns the stored group, version, and kind of an object, or an empty struct
	// if the object does not expose or provide these fields.
	GroupVersionKind() GroupVersionKind
}
```


```go
// /Users/python/go/pkg/mod/k8s.io/apimachinery@v0.24.0/pkg/apis/meta/v1/types.go
type Type interface {
	GetAPIVersion() string
	SetAPIVersion(version string)
	GetKind() string
	SetKind(kind string)
}
```

```go
func (obj *TypeMeta) GetObjectKind() schema.ObjectKind { return obj }

// SetGroupVersionKind satisfies the ObjectKind interface for all objects that embed TypeMeta
func (obj *TypeMeta) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}

// GroupVersionKind satisfies the ObjectKind interface for all objects that embed TypeMeta
func (obj *TypeMeta) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}
```
metav1.TypeMeta 实现了 schema.ObjectKind 接口，schema.ObjectKind 处理所有序列化对象怎么解码与编码资源类型信息的方法


### metav1.ObjectMeta

metav1.ObjectMeta 则用来定义资源对象实例的属性，即所有资源对象都应该具备的属性。这部分就是和对象本身相关，和类型无关，所以相同类型的资源对象的 metav1.ObjectMeta 可能是不同的。

```go
type Object interface {
	GetNamespace() string
	SetNamespace(namespace string)
	GetName() string
	SetName(name string)
	GetGenerateName() string
	SetGenerateName(name string)
	GetUID() types.UID
	SetUID(uid types.UID)
	GetResourceVersion() string
	SetResourceVersion(version string)
	GetGeneration() int64
	SetGeneration(generation int64)
	GetSelfLink() string
	SetSelfLink(selfLink string)
	GetCreationTimestamp() Time
	SetCreationTimestamp(timestamp Time)
	GetDeletionTimestamp() *Time
	SetDeletionTimestamp(timestamp *Time)
	GetDeletionGracePeriodSeconds() *int64
	SetDeletionGracePeriodSeconds(*int64)
	GetLabels() map[string]string
	SetLabels(labels map[string]string)
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
	GetFinalizers() []string
	SetFinalizers(finalizers []string)
	GetOwnerReferences() []OwnerReference
	SetOwnerReferences([]OwnerReference)
	GetZZZ_DeprecatedClusterName() string
	SetZZZ_DeprecatedClusterName(clusterName string)
	GetManagedFields() []ManagedFieldsEntry
	SetManagedFields(managedFields []ManagedFieldsEntry)
}

type ObjectMetaAccessor interface {
    GetObjectMeta() Object
}
```

metav1.ObjectMeta 还实现了 metav1.Object 与 metav1.MetaAccessor 这两个接口
```go
func (obj *ObjectMeta) GetObjectMeta() Object { return obj }

// Namespace implements metav1.Object for any object with an ObjectMeta typed field. Allows
// fast, direct access to metadata fields for API objects.
func (meta *ObjectMeta) GetNamespace() string                { return meta.Namespace }
func (meta *ObjectMeta) SetNamespace(namespace string)       { meta.Namespace = namespace }
func (meta *ObjectMeta) GetName() string                     { return meta.Name }
func (meta *ObjectMeta) SetName(name string)                 { meta.Name = name }
func (meta *ObjectMeta) GetGenerateName() string             { return meta.GenerateName }
func (meta *ObjectMeta) SetGenerateName(generateName string) { meta.GenerateName = generateName }
// ... 
```


### metav1.ListMeta
```go
type ListInterface interface {
	GetResourceVersion() string
	SetResourceVersion(version string)
	GetSelfLink() string
	SetSelfLink(selfLink string)
	GetContinue() string
	SetContinue(c string)
	GetRemainingItemCount() *int64
	SetRemainingItemCount(c *int64)
}
type ListMetaAccessor interface {
    GetListMeta() ListInterface
}
```

metav1.ListMeta 还实现了 metav1.ListInterface 与 metav1.ListMetaAccessor 这两个接口，其中 metav1.ListInterface 接口定义了获取资源对象列表各种元信息的 Get 与 Set 方法：
```go
func (meta *ListMeta) GetResourceVersion() string        { return meta.ResourceVersion }
func (meta *ListMeta) SetResourceVersion(version string) { meta.ResourceVersion = version }
func (meta *ListMeta) GetSelfLink() string               { return meta.SelfLink }
func (meta *ListMeta) SetSelfLink(selfLink string)       { meta.SelfLink = selfLink }
func (meta *ListMeta) GetContinue() string               { return meta.Continue }
func (meta *ListMeta) SetContinue(c string)              { meta.Continue = c }
func (meta *ListMeta) GetRemainingItemCount() *int64     { return meta.RemainingItemCount }
func (meta *ListMeta) SetRemainingItemCount(c *int64)    { meta.RemainingItemCount = c }

func (obj *ListMeta) GetListMeta() ListInterface { return obj }
```

### runtime.Object
schema.ObjectKind 是所有 API 资源类型的抽象，metav1.Object 是所有 API 单体资源对象属性的抽象


那么同时实现这两个接口的类型对象不就可以访问任何 API 对象的公共属性了吗？是的，对于每一个特定的类型，如 Pod、Deployment 等，它们确实可以获取当前 API 对象的公共属性。有没有一种所有特定类型的统一父类，同时拥有 schema.ObjecKind 和 metav1.Object 两个接口，这样就可以表示任何特定类型的对象
```go
// /Users/python/go/pkg/mod/k8s.io/apimachinery@v0.24.0/pkg/runtime/interfaces.go
type Object interface {
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() Object
}
```
为什么 runtime.Object 接口只有这两个方法，不应该有 GetObjectMeta() 方法来获取 metav1.ObjectMeta 对象吗？

```go
// /Users/python/go/pkg/mod/k8s.io/apimachinery@v0.24.0/pkg/api/meta/meta.go
func Accessor(obj interface{}) (metav1.Object, error) {
	switch t := obj.(type) {
	case metav1.Object:
		return t, nil
	case metav1.ObjectMetaAccessor:
		if m := t.GetObjectMeta(); m != nil {
			return m, nil
		}
		return nil, errNotObject
	default:
		return nil, errNotObject
	}
}
```
这样避免了每个 API 资源类型都需要实现 GetObjectMeta() 方法了。

API 资源类型实现 runtime.Object.DeepCopyObject() 方法, 深拷贝方法是具体 API 资源类型需要重载实现的，存在类型依赖，作为 API 资源类型的父类不能统一实现。
一般来说，深拷贝方法是由工具自动生成的，定义在 zz_generated.deepcopy.go 文件中，以 configMap 为例：
```go
// /Users/python/go/pkg/mod/k8s.io/api@v0.24.0/core/v1/zz_generated.deepcopy.go
// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigMap.
func (in *ConfigMap) DeepCopy() *ConfigMap {
	if in == nil {
		return nil
	}
	out := new(ConfigMap)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConfigMap) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
```
