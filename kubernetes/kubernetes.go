package kubernetes

import (
	"context"
	"flag"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func CreateClientSet() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func GetServices(c *kubernetes.Clientset) ([]v1.Service, error) {
	listOptions := metav1.ListOptions{}
	serviceList, err := c.CoreV1().Services("").List(context.Background(), listOptions)

	if err != nil {
		return nil, err
	}

	services := serviceList.Items

	loadBalancerServices := []v1.Service{}

	for _, service := range services {
		if service.Spec.Type == v1.ServiceTypeLoadBalancer {
			loadBalancerServices = append(loadBalancerServices, service)
		}
	}

	return loadBalancerServices, nil
}
