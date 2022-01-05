package main

import (
	"github.com/Erik142/routeros-auto-port-forward/kubernetes"
	"os"
	"os/signal"
	"syscall"
)

func cleanup() {
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
	kubernetes.Listen()
}
