package main

import (
	"fmt"
	"github.com/Erik142/routeros-auto-port-forward/kubernetes"
	"os"
)

func main() {
	clientSet, err := kubernetes.CreateClientSet()

	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(-1)
	}

	services, err := kubernetes.GetServices(clientSet)

	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(-1)
	}

	for _, service := range services {
		fmt.Printf("%s, %s\n", service.Name, service.Spec.Type)
	}
}
