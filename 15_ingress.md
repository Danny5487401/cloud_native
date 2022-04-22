# ingress

阿里云称之为ingress路由！在Kubernetes集群中,主要用于接入外部请求到k8s内部，Ingress是授权入站连接到达集群服务的规则集合，为您提供七层负载均衡能力。您可以给 Ingress 配置提供外部可访问的URL、负载均衡、SSL、基于名称的虚拟主机等。

## 背景
单独用service暴露服务的方式，在实际生产环境中不太合适

- ClusterIP的方式只能在集群内部访问。
- NodePort方式的话，测试环境使用还行，当有几十上百的服务在集群中运行时，NodePort的端口管理是灾难。
- LoadBalance方式受限于云平台，且通常在云平台部署ELB还需要额外的费用。


## ingress与ingress-controller

- ingress对象： 指的是k8s中的一个api对象，一般用yaml配置。作用是定义请求如何转发到service的规则，可以理解为配置模板。
- ingress-controller： 具体实现反向代理及负载均衡的程序，对ingress定义的规则进行解析，根据配置的规则来实现请求转发。

## ingress-controller
ingress-controller并不是k8s自带的组件，实际上ingress-controller只是一个统称，用户可以选择不同的ingress-controller实现，
目前，由k8s维护的ingress-controller只有google云的GCE与ingress-nginx两个，其他还有很多第三方维护的ingress-controller.

一般来说，ingress-controller的形式都是一个pod，里面跑着daemon程序和反向代理程序。daemon负责不断监控集群的变化，根据ingress对象生成配置并应用新配置到反向代理，比如nginx-ingress就是动态生成nginx配置，动态更新upstream，并在需要的时候reload程序应用新配置。


## ingress
ingress是一个API对象，和其他对象一样，通过yaml文件来配置。ingress通过http或https暴露集群内部service，给service提供外部URL、负载均衡、SSL/TLS能力以及基于host的方向代理。

note: extensions/v1beta1 和 networking.k8s.io/v1beta1 API 版本的 Ingress 不在 v1.22 版本中继续提供。
```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: abc-ingress
  annotations: 
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  tls:
  - hosts:
    - api.abc.com
    secretName: abc-tls
  rules:
  - host: api.abc.com
    http:
      paths:
      - backend:
          serviceName: apiserver
          servicePort: 80
  - host: www.abc.com
    http:
      paths:
      - path: /image/*
        backend:
          serviceName: fileserver
          servicePort: 80
  - host: www.abc.com
    http:
      paths:
      - backend:
          serviceName: feserver
          servicePort: 8080

```


## ingress的部署

ingress的部署，需要考虑两个方面：

1. ingress-controller是作为pod来运行的，以什么方式部署比较好
2. ingress解决了把如何请求路由到集群内部，那它自己怎么暴露给外部比较好

### Deployment+LoadBalancer模式的Service

如果要把ingress部署在公有云，那用这种方式比较合适。用Deployment部署ingress-controller，创建一个type为LoadBalancer的service关联这组pod。
大部分公有云，都会为LoadBalancer的service自动创建一个负载均衡器，通常还绑定了公网地址。只要把域名解析指向该地址，就实现了集群服务的对外暴露。


### Deployment+NodePort模式的Service
用deployment模式部署ingress-controller，并创建对应的服务，但是type为NodePort。这样，ingress就会暴露在集群节点ip的特定端口上。由于nodeport暴露的端口是随机端口，一般会在前面再搭建一套负载均衡器来转发请求。该方式一般用于宿主机是相对固定的环境ip地址不变的场景。
NodePort方式暴露ingress虽然简单方便，但是NodePort多了一层NAT，在请求量级很大时可能对性能会有一定影响。

### DaemonSet+HostNetwork+nodeSelector
用DaemonSet结合nodeselector来部署ingress-controller到特定的node上，然后使用HostNetwork直接把该pod与宿主机node的网络打通，直接使用宿主机的80/433端口就能访问服务。
这时，ingress-controller所在的node机器就很类似传统架构的边缘节点，比如机房入口的nginx服务器。
该方式整个请求链路最简单，性能相对NodePort模式更好。
缺点是由于直接利用宿主机节点的网络和端口，一个node只能部署一个ingress-controller pod。比较适合大并发的生产环境使用。