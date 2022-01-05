package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Erik142/routeros-auto-port-forward/kubernetes"

	"github.com/go-routeros/routeros"
	"github.com/pborman/getopt"
)

const BaseAddPortForwardCommand = "/ip/firewall/nat/add,=chain=dstnat,=dst-port=%d,=action=dst-nat,=protocol=%s,=to-addresses=%s,=to-ports=%d,=comment=%s"

const BaseDeletePortForwardCommand = "/ip/firewall/nat/remove,=.id=%s"

const BaseGetPortForwardsCommand = "/ip/firewall/nat/getall,?comment=%s,=.proplist=.id"

func GetAddPortForwardCommand(destinationIp string, destinationPort int, protocol string, comment string) []string {
	return strings.Split(fmt.Sprintf(BaseAddPortForwardCommand, destinationPort, protocol, destinationIp, destinationPort, comment), ",")
}

func GetDeletePortForwardCommand(comment string) []string {
	return strings.Split(fmt.Sprintf(BaseDeletePortForwardCommand, comment), ",")
}

func GetPortForwardsCommand(comment string) []string {
	return strings.Split(fmt.Sprintf(BaseGetPortForwardsCommand, comment), ",")
}

func main() {
	address := getopt.StringLong("address", 'a', "localhost:8728", "The complete URL to the mikrotik router, for example: localhost:8728")
	username := getopt.StringLong("user", 'u', "admin", "The username to login with on the MikroTik router.")
	password := getopt.StringLong("password", 'p', "password", "The password for the router that will be used to login on the MikroTik router.")
	getopt.Parse()

	client, err := routeros.Dial(*address, *username, *password)

	if err != nil {
		log.Printf("%s", err)
		os.Exit(-1)
	}

	clientSet, err := kubernetes.CreateClientSet()

	if err != nil {
		log.Printf("%s", err)
		os.Exit(-1)
	}

	services, err := kubernetes.GetServices(clientSet)

	if err != nil {
		log.Printf("%s", err)
		os.Exit(-1)
	}

	for _, service := range services {
		log.Printf("%s, %s\n", service.Name, service.Spec.Type)
	}

	portForwardCommand := GetAddPortForwardCommand("10.20.30.40", 1234, "tcp", "This is a comment")
	log.Println("Command:", portForwardCommand)
	reply, err := client.RunArgs(portForwardCommand)

	if err != nil {
		log.Printf("%s", err)
		os.Exit(-1)
	}

	log.Println(reply)

	time.Sleep(5 * time.Second)

	getPortForwardsCommand := GetPortForwardsCommand("This is a comment")
	log.Println("Command:", getPortForwardsCommand)
	reply, err = client.RunArgs(getPortForwardsCommand)

	if err != nil {
		log.Printf("%s", err)
		os.Exit(-1)
	}

	for _, sentence := range reply.Re {
		id := sentence.Map[".id"]
		deletePortForwardCommand := GetDeletePortForwardCommand(id)

		reply, err = client.RunArgs(deletePortForwardCommand)

		if err != nil {
			log.Printf("%s", err)
			os.Exit(-1)
		}
	}

	log.Println(reply)
	/*
		deletePortForwardCommand := GetDeletePortForwardCommand("This is a comment")
		log.Println("Command:", deletePortForwardCommand)
		reply, err = client.RunArgs(deletePortForwardCommand)

		if err != nil {
			log.Printf("%s", err)
			os.Exit(-1)
		}

		log.Println(reply)
	*/

	client.Close()
}
