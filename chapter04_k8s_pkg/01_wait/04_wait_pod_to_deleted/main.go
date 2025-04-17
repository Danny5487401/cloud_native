package main

import (
	"context"
	"flag"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"
)

func main() {
	waitClean()

}

func waitClean() error {
	var kubeconfig *string

	// home是家目录，如果能取得家目录的值，就可以用来做默认值
	if home := homedir.HomeDir(); home != "" {
		// 如果输入了kubeconfig参数，该参数的值就是kubeconfig文件的绝对路径，
		// 如果没有输入kubeconfig参数，就用默认路径~/.kube/config
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		// 如果取不到当前用户的家目录，就没办法设置kubeconfig的默认目录了，只能从入参中取
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// 从本机加载kubeconfig配置文件，因此第一个参数为空字符串
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Error(err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err.Error())
	}

	return wait.PollImmediate(
		2*time.Second,
		90*time.Second,
		func() (done bool, err error) {
			_, err = kubeClient.CoreV1().Pods("namespace").Get(context.Background(), "name", metav1.GetOptions{ResourceVersion: "0"})
			if err != nil {
				if errors.IsNotFound(err) {
					return true, nil
				}
				log.Errorf("wait pod error: %s", err.Error())
				return false, nil
			}
			err = kubeClient.CoreV1().Pods("namespace").Delete(context.Background(), "name", &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("delete pod error: %s", err.Error())
				return false, nil
			}
			return false, nil
		},
	)
}
