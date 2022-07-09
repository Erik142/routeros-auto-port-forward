package mikrotik

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-routeros/routeros"
	"github.com/go-routeros/routeros/proto"
)

const BaseAddPortForwardCommand = "/ip/firewall/nat/add|=chain=dstnat|=dst-port=%d|=action=dst-nat|=protocol=%s|=to-addresses=%s|=to-ports=%d|=comment=%s|=in-interface-list=WAN"
const BaseDeletePortForwardCommand = "/ip/firewall/nat/remove|=.id=%s"

const BaseGetPortForwardsCommand = "/ip/firewall/nat/getall|?comment=%s|=.proplist=.id,comment,dst-port,to-addresses,to-ports"
const BaseGetPortForwardsUnequalCommand = "/ip/firewall/nat/getall|?>comment=%s|=.proplist=.id,comment,dst-port,to-addresses,to-ports"

const PortForwardPrefix = "routeros.autoport"

type PortForward struct {
	Namespace       string
	Name            string
	DestinationPort int
	ToPort          int
	DestinationIp   string
	Protocol        string
}

func GetAddPortForwardCommand(destinationIp string, destinationPort int, routerPort int, protocol string, comment string) []string {
	return strings.Split(fmt.Sprintf(BaseAddPortForwardCommand, routerPort, protocol, destinationIp, destinationPort, comment), "|")
}

func GetDeletePortForwardCommand(comment string) []string {
	return strings.Split(fmt.Sprintf(BaseDeletePortForwardCommand, comment), "|")
}

func GetPortForwardsCommand(comment string, equal bool) []string {
	if equal {
		return strings.Split(fmt.Sprintf(BaseGetPortForwardsCommand, comment), "|")
	} else {
		return strings.Split(fmt.Sprintf(BaseGetPortForwardsUnequalCommand, comment), "|")
	}
}

func AddPortForward(client routeros.Client, portForward PortForward) (bool, error) {
	portForwardCommand := GetAddPortForwardCommand(portForward.DestinationIp, portForward.DestinationPort, portForward.DestinationPort, portForward.Protocol, fmt.Sprintf("%s.%s.%s", PortForwardPrefix, portForward.Namespace, portForward.Name))
	log.Println(portForwardCommand)
	_, err := client.RunArgs(portForwardCommand)

	if err != nil {
		log.Fatal(err)
		return false, err
	}

	return true, nil
}

func DeletePortForward(client routeros.Client, portForward PortForward) (bool, error) {
	sentences, err := getPortForwards(client, portForward)

	if err != nil {
		log.Fatal(err)
		return false, err
	}

	for _, sentence := range sentences {
		id := sentence.Map[".id"]
		deletePortForwardCommand := GetDeletePortForwardCommand(id)

		log.Println(deletePortForwardCommand)

		_, err := client.RunArgs(deletePortForwardCommand)

		if err != nil {
			log.Fatal(err)
			return false, err
		}
	}

	return true, nil
}

func GetAllPortForwards(client routeros.Client) ([]PortForward, error) {
	sentences, err := getPortForwards(client, PortForward{})

	if err != nil {
		return nil, err
	}

	portForwards := []PortForward{}

	for _, sentence := range sentences {
		comment, ok := sentence.Map["comment"]

		if ok {
			data := strings.ReplaceAll(comment, fmt.Sprintf("%s.", PortForwardPrefix), "")
			namespace := strings.SplitN(data, ".", 2)[0]
			name := strings.SplitN(data, ".", 2)[1]
			destinationPort, _ := strconv.Atoi(sentence.Map["dst-port"])
			toPort, _ := strconv.Atoi(sentence.Map["to-port"])

			portForwards = append(portForwards, PortForward{
				Namespace:       namespace,
				Name:            name,
				DestinationIp:   sentence.Map["to-addresses"],
				DestinationPort: destinationPort,
				ToPort:          toPort,
			})
		}

	}
	return portForwards, nil
}

func getPortForwards(client routeros.Client, portForward PortForward) ([]*proto.Sentence, error) {
	var getPortForwardsCommand []string
	if portForward.Name == "" && portForward.Namespace == "" {
		getPortForwardsCommand = GetPortForwardsCommand(fmt.Sprintf("%s", PortForwardPrefix), false)
	} else {
		getPortForwardsCommand = GetPortForwardsCommand(fmt.Sprintf("%s.%s.%s", PortForwardPrefix, portForward.Namespace, portForward.Name), true)
	}

	reply, err := client.RunArgs(getPortForwardsCommand)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return reply.Re, nil
}
