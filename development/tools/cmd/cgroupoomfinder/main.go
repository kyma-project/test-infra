package main

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	log "github.com/sirupsen/logrus"
	"sync"
)

var (
	client *containerd.Client
	wg     sync.WaitGroup
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
	wg.Add(2)
	eventService := client.EventService()
	events, errs := eventService.Subscribe(ctx)
	oom := namespaces.WithNamespace(context.Background(), "oom")
	go func() {
		defer wg.Done()
		for {
			cherr, ok := <-errs
			if !ok {
				log.WithFields(log.Fields{"msg": "failed read containerd event error"}).Errorf("%+v", errs)
				return
			}
			log.WithFields(log.Fields{"msg": "got containerd events channel error"}).Errorf("%+v", cherr)
		}
	}()
	go func() {
		defer wg.Done()
		for {
			event, ok := <-events
			if !ok {
				log.Error("failed read containerd event")
			}
			fmt.Printf("%+v", event)
			allContainers, err := client.Containers(oom)
			if err != nil {
				log.WithFields(log.Fields{"msg": "failed get containers"}).Fatalf("error: %s", err)
			}
			for container := range allContainers {
				fmt.Printf("%+v", container)
			}
		}
	}()
	wg.Wait()
}
