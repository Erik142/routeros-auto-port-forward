package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Erik142/routeros-auto-port-forward/kubernetes"
)

func cleanup() {
	log.Println("Cleaning up...")
	kubernetes.Close()
}

func main() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	log.Println("Listening for service changes...")
	kubernetes.Listen()
}
