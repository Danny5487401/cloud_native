<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [k8s基本网络模型](#k8s%E5%9F%BA%E6%9C%AC%E7%BD%91%E7%BB%9C%E6%A8%A1%E5%9E%8B)
  - [underlay](#underlay)
    - [1. 大二层网络（node和pod在同一个网段）](#1-%E5%A4%A7%E4%BA%8C%E5%B1%82%E7%BD%91%E7%BB%9Cnode%E5%92%8Cpod%E5%9C%A8%E5%90%8C%E4%B8%80%E4%B8%AA%E7%BD%91%E6%AE%B5)
    - [2. 大三层网络（node和pod在不同一个网段）](#2-%E5%A4%A7%E4%B8%89%E5%B1%82%E7%BD%91%E7%BB%9Cnode%E5%92%8Cpod%E5%9C%A8%E4%B8%8D%E5%90%8C%E4%B8%80%E4%B8%AA%E7%BD%91%E6%AE%B5)
    - [扩展：中间不是交换机，是路由器(跨网段,可以跨vpc)](#%E6%89%A9%E5%B1%95%E4%B8%AD%E9%97%B4%E4%B8%8D%E6%98%AF%E4%BA%A4%E6%8D%A2%E6%9C%BA%E6%98%AF%E8%B7%AF%E7%94%B1%E5%99%A8%E8%B7%A8%E7%BD%91%E6%AE%B5%E5%8F%AF%E4%BB%A5%E8%B7%A8vpc)
  - [overlay(隧道模式)](#overlay%E9%9A%A7%E9%81%93%E6%A8%A1%E5%BC%8F)
  - [docker的网络方案](#docker%E7%9A%84%E7%BD%91%E7%BB%9C%E6%96%B9%E6%A1%88)
    - [Docker网络的局限性](#docker%E7%BD%91%E7%BB%9C%E7%9A%84%E5%B1%80%E9%99%90%E6%80%A7)
  - [Netns(network namespace)](#netnsnetwork-namespace)
    - [定义](#%E5%AE%9A%E4%B9%89)
    - [使用](#%E4%BD%BF%E7%94%A8)
    - [Pod 与 Netns 的关系](#pod-%E4%B8%8E-netns-%E7%9A%84%E5%85%B3%E7%B3%BB)
  - [网络设备](#%E7%BD%91%E7%BB%9C%E8%AE%BE%E5%A4%87)
  - [k8s网络模型的原则](#k8s%E7%BD%91%E7%BB%9C%E6%A8%A1%E5%9E%8B%E7%9A%84%E5%8E%9F%E5%88%99)
    - [IP-Per-Pod与Docker端口映射的区别](#ip-per-pod%E4%B8%8Edocker%E7%AB%AF%E5%8F%A3%E6%98%A0%E5%B0%84%E7%9A%84%E5%8C%BA%E5%88%AB)
  - [k8s网络模型](#k8s%E7%BD%91%E7%BB%9C%E6%A8%A1%E5%9E%8B)
    - [1. 容器与容器的通讯](#1-%E5%AE%B9%E5%99%A8%E4%B8%8E%E5%AE%B9%E5%99%A8%E7%9A%84%E9%80%9A%E8%AE%AF)
    - [2. pod与pod的通讯](#2-pod%E4%B8%8Epod%E7%9A%84%E9%80%9A%E8%AE%AF)
      - [同一Node内的pod之间通讯](#%E5%90%8C%E4%B8%80node%E5%86%85%E7%9A%84pod%E4%B9%8B%E9%97%B4%E9%80%9A%E8%AE%AF)
      - [不同Node的pod之间通讯](#%E4%B8%8D%E5%90%8Cnode%E7%9A%84pod%E4%B9%8B%E9%97%B4%E9%80%9A%E8%AE%AF)
  - [IASS 主流网络方案](#iass-%E4%B8%BB%E6%B5%81%E7%BD%91%E7%BB%9C%E6%96%B9%E6%A1%88)
  - [Network Policy](#network-policy)
  - [参考资料](#%E5%8F%82%E8%80%83%E8%B5%84%E6%96%99)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# k8s基本网络模型
![](../img/.08_k8s_network_model_images/k8s_network_model2.png)

分类：根据是否寄生在 Host 网络之上可以把容器网络方案大体分为 Underlay/Overlay 两大派别
    
* Underlay 的标准是它与 Host 网络是同层的，从外在可见的一个特征就是它是不是使用了 Host 网络同样的网段、输入输出基础设备、容器的 IP 地址是不是需要与 Host 网络取得协同（来自同一个中心分配或统一划分）。

* Overlay 不一样的地方就在于它并不需要从 Host 网络的 IPM 的管理的组件去申请IP，一般来说，它只需要跟 Host 网络不冲突，这个 IP 可以自由分配的。

![](.08_k8s_network_model_images/sr-iov_process.png)

SR-IOV（Single Root I/O Virtualization）:Intel 在 2007年提出的一种基于硬件的虚拟化解决方案,支持了单个物理PCIe设备虚拟出多个虚拟PCIe设备，然后将虚拟PCIe设备直通到各虚拟机，以实现单个物理PCIe设备支撑多虚拟机的应用场景

SR-IOV 使用 physical functions (PF) 和 virtual functions (VF) 为 SR-IOV 设备管理全局功能。


## underlay
### 1. 大二层网络（node和pod在同一个网段）
![](.08_k8s_network_model_images/big_2_network.png)
- 主要依据：交换机arp广播获取mac地址
- 背景：pod的Ip在物理世界对于交换机是不认识的，pod1(192.168.1.100)不知道pod2的Ip(192.168.1.101)
- 流程：
    1. 需要配置把pod1网关Ip指向node1(192.168.1.200),通过虚拟网线veth连接虚拟网桥bridge流向node1,
    2. node1的netfilter会因为同网段会进行处理发向交换机,arp给pod2,但是在物理世界不知道pod2虚拟Ip。
    3. 需要软件定义交换机SDN（Software Defined Network）伪造ARP应答:192.168.1.201的mac地址是Node2，这样可以把消息从node1发给node2。
    4. node2的netfilter进入内核，然后通过软件本地路由表，192.168.1.201/32使用veth pair进入po2，
    
- 特点：pod2创建，要让pod1知道，所以bgp要进行下发,即帮你在多个node之间同步路由表,所以node上要有bgp agent。

### 2. 大三层网络（node和pod在不同一个网段）
![](.08_k8s_network_model_images/big_3_network.png)
![](.08_k8s_network_model_images/big_3_network2.png)
- 应用：calico的bgp模式：Border Gateway Protocol
- 流程：
    1. 目标地址172.168.1.x设置网关是node1:192.168.1.100,流量走向node1
    2. 三层路由表：bgp设置172.168.1.101/32发往网关node2:192.168.1.101,从eth0网口送出
    3. arp广播192.168.1.101,node2进行arp回答mac地址，交换机没有进行sdn这种行为。
    4. 过本机路由表192.168.1.101/32发往虚拟网桥bridge172.168.1.x。
    
### 扩展：中间不是交换机，是路由器(跨网段,可以跨vpc)
![](.08_k8s_network_model_images/big_3_network3.png)

## overlay(隧道模式)
![](.08_k8s_network_model_images/overlay_process1.png)
- 优点：对物理网络没有要求
- 缺点：
    - 封装解包，计算量上升，延迟。
    - 隧道使payload增加，在mtu固定1500时，增加头部，对应payload值会减少
    
- 过程：
    1. node1路由规则：172.168.1.201/32发往tun0设备，写real ip是192.168.1.101;然后封包src变成192.168.1.100，des变成192.168.1.101，
    payload是src：172.168.1.200,dst:172.168.1.201,data不变
    2. 经过交换机或则路由器，node2上的eth0进行解包。
    3. 经过netfilter进行forward,本地路由表直接通过veth pair发给对应应用,同网段可以没有bridge。

  


## docker的网络方案
docker官方并没有提供多主机的容器通信方案，单机网络的模式主要有host，container，bridge，none。
![](../img/.08_k8s_network_model_images/docker_network.png)

- none
- host，与宿主机共享，占用宿主机资源
- container，使用某容器的namespace，例如k8s的同一pod内的各个容器
- bridge，挂到网桥docker0上，走iptables做NAT

### Docker网络的局限性
- Docker网络模型没有考虑到多主机互联的网络解决方案，崇尚简单为美
- 同一机器内的容器之间可以直接通讯，但是不同机器之间的容器无法通讯
- 为了跨节点通讯，必须在主机的地址上分配端口，通过端口路由或代理到容器
- 分配和管理容器特别困难，特别是水平扩展时

## Netns(network namespace)
需要了解的内容
![](../img/.08_k8s_network_model_images/netns_menu.png)

### 定义
![](../img/.08_k8s_network_model_images/netns.png)
![](../img/.08_k8s_network_model_images/netns_definition.png)   
网络 由网络接口,iptables,路由表 构成

1. 网卡
![](../img/.08_k8s_network_model_images/network_card.png)

2. iptables
![](../img/.08_k8s_network_model_images/iptables.png)

3. 路由表
![](../img/.08_k8s_network_model_images/route_info.png)

### 使用
![](../img/.08_k8s_network_model_images/trace_route.png)

1. 自己创建netns

![](../img/.08_k8s_network_model_images/add_netns.png)
![](../img/.08_k8s_network_model_images/netns_operator.png)
![](../img/.08_k8s_network_model_images/netns_operator2.png)
![](../img/.08_k8s_network_model_images/netns_operator3.png)
![](../img/.08_k8s_network_model_images/netns_operator4.png)

与docker,k8s对比
![](../img/.08_k8s_network_model_images/netns_vs_docker_n_k8s.png)

2. 两个netns交流
方式一：veth   

![](../img/.08_k8s_network_model_images/two_netns.png)
![](../img/.08_k8s_network_model_images/netns_two.png)

开始搭建梯子🪜，一边一半    
![](../img/.08_k8s_network_model_images/ladder.png)

构造梯子veth    
![](../img/.08_k8s_network_model_images/iplink.png)

放梯子到各自家里  
![](../img/.08_k8s_network_model_images/iplink2.png)

固定梯子    
![](../img/.08_k8s_network_model_images/fix_ladder.png)

启动设备     
![](../img/.08_k8s_network_model_images/up_link.png)

开始拍手     
![](../img/.08_k8s_network_model_images/link_communication.png)

方式二：桥
![](../img/.08_k8s_network_model_images/bridge_comm.png)

建立桥  
![](../img/.08_k8s_network_model_images/add_bridge.png)

建立梯子到王婆  
![](../img/.08_k8s_network_model_images/ladder_bridge.png)

放梯子到各自家里:注意王婆是master,不是单独的namespace   
![](../img/.08_k8s_network_model_images/put_ladder_home.png)

查看master王婆的信息  
![](../img/.08_k8s_network_model_images/master_info.png)

固定西门庆家的梯子就行  
![](../img/.08_k8s_network_model_images/fix_ladder_xmq.png)

激活设备(包括王婆的设备ip link set wangpo up)   
![](../img/.08_k8s_network_model_images/set_link_up1.png)

同理去panjinlian家配置  
![](../img/.08_k8s_network_model_images/pjl2wp_ladder.png)
![](../img/.08_k8s_network_model_images/pjl2wp_link_up.png)

方式三：ipvlan(ip不同，mac相同)-->没有经过数据解封装

![](../img/.08_k8s_network_model_images/ipvlan.png)
- 查看mac地址，其实net1和net2的mac地址一样的。
- 子接口172.12.1.5和子接口172.12.1.6通的
- 子接口172.12.1.5和父接口172.12.1.30不通的

![](../img/.08_k8s_network_model_images/child_n_parent_info.png)
- 子接口172.12.1.5和网关172.12.1.2通的
- 子接口172.12.1.5和电信114.114.114.114不通的

![](../img/.08_k8s_network_model_images/without_route_114.png)
![](../img/.08_k8s_network_model_images/add_route_114.png)




### Pod 与 Netns 的关系
![](../img/.08_k8s_network_model_images/relation_between_pod_and_netns.png)

## 网络设备
![](../img/.08_k8s_network_model_images/iso_protocol.png)

1. hub 集线器

![](../img/.08_k8s_network_model_images/hub.png)

特点  
![](../img/.08_k8s_network_model_images/hub_info.png)
![](../img/.08_k8s_network_model_images/hub_info2.png)
![](../img/.08_k8s_network_model_images/hub_info3.png)

 
2. bridge 网桥  

![](../img/.08_k8s_network_model_images/bridge_device.png)
![](../img/.08_k8s_network_model_images/bridge_device_info.png)
![](../img/.08_k8s_network_model_images/bridge_device_mechanism.png)
注意是第二层：mac地址


3. switch 交换机  
![](../img/.08_k8s_network_model_images/switch_device.png)

这里：可以指二层，有些到三层。
![](../img/.08_k8s_network_model_images/switch_info1.png)
![](../img/.08_k8s_network_model_images/switch_info2.png)

与网桥对比
![](../img/.08_k8s_network_model_images/bridge_vs_switch.png)


4. DHCP(动态主机配置协议) Server
![](../img/.08_k8s_network_model_images/dhcp_process.png)
![](../img/.08_k8s_network_model_images/dhcp_process1.png)

5. NAT Device

路由器
![](../img/.08_k8s_network_model_images/route_device.png)
![](../img/.08_k8s_network_model_images/route_device_info.png)
![](../img/.08_k8s_network_model_images/nat_translate.png)

类型:最常用napt     
![](../img/.08_k8s_network_model_images/net_class.png)
![](../img/.08_k8s_network_model_images/static_nat.png)
![](../img/.08_k8s_network_model_images/pool_nat.png)
![](../img/.08_k8s_network_model_images/napt.png)

## k8s网络模型的原则
- 每个pod都拥有唯一个独立的ip地址，称IP-Per-Pod模型
- 所有pod都在一个可连通的网络环境中
- 不管是否在同一个node，都可以通过ip直接通讯
- pod被看作一台独立的物理机或虚拟机

### IP-Per-Pod与Docker端口映射的区别
docker端口映射到宿主机会引入端口管理的复杂性
docker最终被访问的ip和端口，与提供的不一致，引起配置的复杂性



## k8s网络模型
![](../img/.08_k8s_network_model_images/k8s_network_model_info.png)

### 1. 容器与容器的通讯
- 同一个容器的pod直接共享同一个linux协议栈
- 就像在同一台机器上，可通过localhost访问
- 可类比一个物理机上不同应用程序的情况

### 2. pod与pod的通讯
#### 同一Node内的pod之间通讯
- 同一Node内的pod都是通过veth连接在同一个docker0网桥上，地址段相同，所以可以直接通讯

#### 不同Node的pod之间通讯
- docker0网段与宿主机不在同一个网段，所以不同pod之间的pod不能直接通讯
- 不同node之间通讯只能通过宿主机物理网卡
- 前面说过k8s网络模型需要不同的pod之间能通讯，所以ip不能重复，这就要求k8s部署时要规划好docker0的网段
- 同时，要记录每个pod的ip地址挂在哪个具体的node上
- 为了达到这个目的，有很多开源软件增强了docker和k8s的网络



## IASS 主流网络方案
我们可以把云计算理解成一栋大楼，而这栋楼又可以分为顶楼、中间、低层三大块。那么我们就可以把 Iass(基础设施)、Pass(平台)、Sass(软件)理解成这栋楼的三部分
![](../img/.08_k8s_network_model_images/container_network.png)







## Network Policy
定义：提供了基于策略的网络控制，用于隔离应用并减少攻击面。他使用标签选择器模拟传统的分段网络，并通过策略控制他们之间的流量和外部的流量。
注意：在使用network policy之前
    
* apiserver需要开启extensions/v1beta1/networkpolicies
* 网络插件需要支持networkpolicy

Configuration
![](../img/.08_k8s_network_model_images/configuration.png)




## 参考资料
 
- [ip 命令使用](https://blog.csdn.net/qq_35029061/article/details/125967340)
- [Kubernetes 网络插件详解 – Flannel篇](https://www.infvie.com/ops-notes/kubernetes-cni-flannel.html)
- [揭秘 IPIP 隧道](https://morven.life/posts/networking-3-ipip/)