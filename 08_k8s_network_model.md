# k8s基本网络模型

分类：根据是否寄生在 Host 网络之上可以把容器网络方案大体分为 Underlay/Overlay 两大派别
    
* Underlay 的标准是它与 Host 网络是同层的，从外在可见的一个特征就是它是不是使用了 Host 网络同样的网段、输入输出基础设备、容器的 IP 地址是不是需要与 Host 网络取得协同（来自同一个中心分配或统一划分）。这就是 Underlay；

* Overlay 不一样的地方就在于它并不需要从 Host 网络的 IPM 的管理的组件去申请IP，一般来说，它只需要跟 Host 网络不冲突，这个 IP 可以自由分配的。

## docker的网络方案
docker官方并没有提供多主机的容器通信方案，单机网络的模式主要有host，container，bridge，none。
- none
- host，与宿主机共享，占用宿主机资源
- container，使用某容器的namespace，例如k8s的同一pod内的各个容器
- bridge，挂到网桥docker0上，走iptables做NAT


## Netns(network namespace)

需要了解的内容
![](img/.08_k8s_network_model_images/netns_menu.png)

### 定义
![](img/.08_k8s_network_model_images/netns.png)
![](img/.08_k8s_network_model_images/netns_definition.png)
网络接口，iptables,路由表

1. 网卡
![](img/.08_k8s_network_model_images/network_card.png)

2. iptables
![](img/.08_k8s_network_model_images/iptables.png)

3. 路由表
![](img/.08_k8s_network_model_images/route_info.png)

### 使用
![](img/.08_k8s_network_model_images/trace_route.png)

1. 自己创建netns
![](img/.08_k8s_network_model_images/add_netns.png)
![](img/.08_k8s_network_model_images/netns_operator.png)
![](img/.08_k8s_network_model_images/netns_operator2.png)
![](img/.08_k8s_network_model_images/netns_operator3.png)
![](img/.08_k8s_network_model_images/netns_operator4.png)

与docker,k8s对比
![](img/.08_k8s_network_model_images/netns_vs_docker_n_k8s.png)

2. 两个netns交流
方式一：veth
![](img/.08_k8s_network_model_images/two_netns.png)
![](img/.08_k8s_network_model_images/netns_two.png)
开始搭建梯子🪜，一边一半
![](img/.08_k8s_network_model_images/ladder.png)
构造梯子veth
![](img/.08_k8s_network_model_images/iplink.png)
放梯子到各自家里
![](img/.08_k8s_network_model_images/iplink2.png)
固定梯子
![](img/.08_k8s_network_model_images/fix_ladder.png)
启动设备
![](img/.08_k8s_network_model_images/up_link.png)
开始拍手
![](img/.08_k8s_network_model_images/link_communication.png)

方式二：桥
![](img/.08_k8s_network_model_images/bridge_comm.png)
建立桥
![](img/.08_k8s_network_model_images/add_bridge.png)
建立梯子到王婆
![](img/.08_k8s_network_model_images/ladder_bridge.png)
放梯子到各自家里:注意王婆是master,不是单独的namespace 
![](img/.08_k8s_network_model_images/put_ladder_home.png)
查看master王婆的信息
![](img/.08_k8s_network_model_images/master_info.png)
固定西门庆家的梯子就行
![](img/.08_k8s_network_model_images/fix_ladder_xmq.png)
激活设备(包括王婆的设备ip link set wangpo up)
![](img/.08_k8s_network_model_images/set_link_up1.png)

同理去panjinlian家配置
![](img/.08_k8s_network_model_images/pjl2wp_ladder.png)
![](img/.08_k8s_network_model_images/pjl2wp_link_up.png)


### Pod 与 Netns 的关系
![](img/.08_k8s_network_model_images/relation_between_pod_and_netns.png)


## 网络设备
![](img/.08_k8s_network_model_images/iso_protocol.png)
1. hub 集线器
![](img/.08_k8s_network_model_images/hub.png)
特点
![](img/.08_k8s_network_model_images/hub_info.png)
![](img/.08_k8s_network_model_images/hub_info2.png)
![](img/.08_k8s_network_model_images/hub_info3.png)


2. bridge 网桥
![](img/.08_k8s_network_model_images/bridge_device.png)
![](img/.08_k8s_network_model_images/bridge_device_info.png)
![](img/.08_k8s_network_model_images/bridge_device_mechanism.png)
注意是第二层：mac地址


3. switch 交换机
![](img/.08_k8s_network_model_images/switch_device.png)
这里：可以指二层，有些到三层。
![](img/.08_k8s_network_model_images/switch_info1.png)
![](img/.08_k8s_network_model_images/switch_info2.png)

与网桥对比
![](img/.08_k8s_network_model_images/bridge_vs_switch.png)


4. DHCP(动态主机配置协议) Server
![](img/.08_k8s_network_model_images/dhcp_process.png)
![](img/.08_k8s_network_model_images/dhcp_process1.png)

5. NAT Device
路由器
![](img/.08_k8s_network_model_images/route_device.png)
![](img/.08_k8s_network_model_images/route_device_info.png)
![](img/.08_k8s_network_model_images/nat_translate.png)

类型:最常用napt
![](img/.08_k8s_network_model_images/net_class.png)
![](img/.08_k8s_network_model_images/static_nat.png)
![](img/.08_k8s_network_model_images/pool_nat.png)
![](img/.08_k8s_network_model_images/napt.png)

## 主流网络方案
我们可以把云计算理解成一栋大楼，而这栋楼又可以分为顶楼、中间、低层三大块。那么我们就可以把Iass(基础设施)、Pass(平台)、Sass(软件)理解成这栋楼的三部分
![](img/.08_k8s_network_model_images/container_network.png)

### Flannel
![](img/.08_k8s_network_model_images/flannel.png)

它首先要解决的是 container 的包如何到达 Host，这里采用的是加一个 Bridge 的方式。
它的 backend 其实是独立的，也就是说这个包如何离开 Host，是采用哪种封装方式，还是不需要封装，都是可选择的

三种主要的 backend：

* 一种是用户态的 udp，这种是最早期的实现；
* 然后是内核的 Vxlan，这两种都算是 overlay 的方案。Vxlan 的性能会比较好一点，但是它对内核的版本是有要求的，需要内核支持 Vxlan 的特性功能；
* 如果你的集群规模不够大，又处于同一个二层域，也可以选择采用 host-gw 的方式。这种方式的 backend 基本上是由一段广播路由规则来启动的，性能比较高

#### Flannel的大致流程
1. flannel利用Kubernetes API或者etcd用于存储整个集群的网络配置，其中最主要的内容为设置集群的网络地址空间。例如，设定整个集群内所有容器的IP都取自网段“10.1.0.0/16”。
2. flannel在每个主机中运行flanneld作为agent，它会为所在主机从集群的网络地址空间中，获取一个小的网段subnet，本主机内所有容器的IP地址都将从中分配。
3. flanneld再将本主机获取的subnet以及用于主机间通信的Public IP，同样通过kubernetes API或者etcd存储起来。
4. flannel利用各种backend ，例如udp，vxlan，host-gw等等，跨主机转发容器间的网络流量，完成容器间的跨主机通信。



#### Flannel的设置方式
Flanneld是Flannel守护程序，通常作为守护程序安装在kubernetes集群上，并以install-cni作为初始化容器。
install-cni容器在每个节点上创建CNI配置文件-/etc/cni/net.d/10-flannel.conflist。
Flanneld创建一个vxlan设备，从apiserver获取网络元数据，并监视pod上的更新。
创建Pod时，它将为整个集群中的所有Pod分配路由，这些路由允许Pod通过其IP地址相互连接。

![](img/.08_k8s_network_model_images/cri_n_cni.png)

kubelet调用Containered CRI插件以创建容器，而Containered CRI插件调用CNI插件为容器配置网络。
网络提供商CNI插件调用其他基本CNI插件来配置网络。

## Network Policy
定义：提供了基于策略的网络控制，用于隔离应用并减少攻击面。他使用标签选择器模拟传统的分段网络，并通过策略控制他们之间的流量和外部的流量。
注意：在使用network policy之前
    
* apiserver需要开启extensions/v1beta1/networkpolicies
* 网络插件需要支持networkpolicy

Configuration
![](img/.08_k8s_network_model_images/configuration.png)
    