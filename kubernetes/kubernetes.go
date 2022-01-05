package kubernetes

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Erik142/routeros-auto-port-forward/mikrotik"
	"github.com/go-routeros/routeros"
	"github.com/pborman/getopt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const PortForwardEnabledAnnotation = "routeros.portforward.enabled"
const PortForwardPorts = "routeros.portforward.ports"

var routerOsClient *routeros.Client

func createClientSet() (*kubernetes.Clientset, error) {
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

func getServices(c *kubernetes.Clientset) ([]v1.Service, error) {
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

func Close() {
	routerOsClient.Close()
}

func Listen() {
	address := getopt.StringLong("address", 'a', "localhost:8728", "The complete URL to the mikrotik router, for example: localhost:8728")
	username := getopt.StringLong("user", 'u', "admin", "The username to login with on the MikroTik router.")
	password := getopt.StringLong("password", 'p', "password", "The password for the router that will be used to login on the MikroTik router.")
	getopt.Parse()

	kubernetesClientSet, err := createClientSet()

	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	routerOsClient, err = routeros.Dial(*address, *username, *password)

	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	for {
		services, err := getServices(kubernetesClientSet)

		if err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}

		portForwards, err := mikrotik.GetAllPortForwards(*routerOsClient)

		if err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}

		oldPortForwards, err := getOldPortForwards(portForwards, services)

		if err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}

		newPortForwards, err := getNewPortForwards(portForwards, services)

		if err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}

		if oldPortForwards != nil && len(oldPortForwards) > 0 {
			for _, portForward := range oldPortForwards {
				log.Printf("Deleting the port forward for the service %s in the namespace %s\n", portForward.Name, portForward.Namespace)
				mikrotik.DeletePortForward(*routerOsClient, portForward)
			}
		}

		if newPortForwards != nil && len(newPortForwards) > 0 {
			for _, portForward := range newPortForwards {
				log.Printf("Adding a new port forward for the service %s in the namespace %s\n", portForward.Name, portForward.Namespace)
				mikrotik.AddPortForward(*routerOsClient, portForward)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func getOldPortForwards(allPortForwards []mikrotik.PortForward, allServices []v1.Service) ([]mikrotik.PortForward, error) {
	oldPortForwards := []mikrotik.PortForward{}

	if len(allServices) == 0 {
		return allPortForwards, nil
	}

	for _, portForward := range allPortForwards {
		index := sort.Search(len(allServices), func(i int) bool {
			return allServices[i].Namespace == portForward.Namespace && allServices[i].Status.LoadBalancer.Ingress[0].IP == portForward.DestinationIp && allServices[i].Name == portForward.Name
		})

		if index == len(allServices) {
			oldPortForwards = append(oldPortForwards, portForward)
		}
	}

	return oldPortForwards, nil
}

func getNewPortForwards(allPortForwards []mikrotik.PortForward, allServices []v1.Service) ([]mikrotik.PortForward, error) {
	newPortForwards := []mikrotik.PortForward{}

	for _, service := range allServices {
		index := sort.Search(len(allPortForwards), func(i int) bool {
			return allPortForwards[i].Namespace == service.Namespace && allPortForwards[i].DestinationIp == service.Status.LoadBalancer.Ingress[0].IP && allPortForwards[i].Name == service.Name
		})

		if index == len(allPortForwards) && len(service.Status.LoadBalancer.Ingress) > 0 && service.Status.LoadBalancer.Ingress[0].IP != "" {
			for _, servicePort := range service.Spec.Ports {
				destinationPort := int(servicePort.Port)
				toPort := int(servicePort.Port)

				newPortForwards = append(newPortForwards, mikrotik.PortForward{
					Namespace:       service.Namespace,
					Name:            service.Name,
					DestinationPort: destinationPort,
					DestinationIp:   service.Status.LoadBalancer.Ingress[0].IP,
					ToPort:          toPort,
					Protocol:        strings.ToLower(string(servicePort.Protocol)),
				})
			}
		}
	}

	return newPortForwards, nil
}
