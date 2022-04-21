package main


import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// 从namespace为kube-system中获取所有的pod并输出到屏幕

func main() {
	fmt.Println("Prepare config object.")

	// 加载k8s配置文件，生成Config对象
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		panic(err)
	}

	config.APIPath = "api"
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs

	fmt.Println("Init RESTClient.")

	// 定义RestClient，用于与k8s API server进行交互
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}

	fmt.Println("Get Pods in cluster.")

	// 获取pod列表。这里只会从namespace为"kube-system"中获取指定的资源(pods)
	result := &corev1.PodList{}
	if err := restClient.
		Get().
		Namespace("kube-system").
		Resource("pods").
		VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(result); err != nil {
		panic(err)
	}

	fmt.Println("Print all listed pods.")

	// 打印所有获取到的pods资源，输出到标准输出
	for _, d := range result.Items {
		fmt.Printf("NAMESPACE: %v NAME: %v \t STATUS: %v \n", d.Namespace, d.Name, d.Status.Phase)
	}
}