package main

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	log "github.com/sirupsen/logrus"
)

var (
	client *containerd.Client
)

func init() {
	var err error
	client, err = containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		log.WithFields(log.Fields{"msg": "failed create containerd client"}).Fatalf("error: %s", err)
	}
}

func main() {
	defer client.Close()
	ctx := context.Background()
	eventService := client.EventService()
	event, cherr := eventService.Subscribe(ctx)
	if cherr != nil {
		log.WithFields(log.Fields{"msg": "failed read containerd event"}).Errorf("%+v", cherr)
	}
	fmt.Printf("%+v", event)
	allContainers, err := client.Containers(ctx)
	if err != nil {
		log.WithFields(log.Fields{"msg": "failed get containers"}).Fatalf("error: %s", err)
	}
	for container := range allContainers {
		fmt.Printf("%+v", container)
	}
}
