package kubernetes

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/pborman/getopt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const PortForwardEnabledAnnotation = "routeros.portforward.enabled"
const PortForwardPorts = "routeros.portforward.ports"

func CreateClientSet() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = getopt.StringLong("kubeconfig", 'c', filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = getopt.StringLong("kubeconfig", 'c', "", "absolute path to the kubeconfig file")
	}
	getopt.Parse()

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
		portForwardEnabledStr, ok := service.Annotations[PortForwardEnabledAnnotation]
		if service.Spec.Type == v1.ServiceTypeLoadBalancer && ok && strings.TrimSpace(strings.ToLower(portForwardEnabledStr)) == "true" {
			loadBalancerServices = append(loadBalancerServices, service)
		}
	}

	return loadBalancerServices, nil
}
