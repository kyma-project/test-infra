package main

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	dockerclient "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

// dockerOOMListener create channel to get events from docker daemon.
// Listening is done in a goroutine.
// Channel will receive only oom events as defined in Filters property of EventsOptions.
// On oom event details are printed to stdout.
func dockerOOMListener(client *dockerclient.Client, wg *sync.WaitGroup) {
	eventsChannel := make(chan *dockerclient.APIEvents)
	eventOptions := dockerclient.EventsOptions{
		Filters: map[string][]string{
			"type":  {"container"},
			"event": {"oom"},
		},
	}
	errs := client.AddEventListenerWithOptions(eventOptions, eventsChannel)
	if errs != nil {
		log.WithFields(log.Fields{"msg": "got docker daemon events error"}).Errorf("%s", errs)
	}
	go func() {
		defer wg.Done()
		for {
			event, ok := <-eventsChannel
			if !ok {
				log.Error("failed read docker event")
			}
			fmt.Fprintf(os.Stdout, "OOM event received time: %s , namespace: %s , pod: %s ,container: %s , image: %s \n",
				time.Unix(event.Time, 0).Format(time.RFC822Z),
				event.Actor.Attributes["io.kubernetes.pod.namespace"],
				event.Actor.Attributes["io.kubernetes.pod.name"],
				event.Actor.Attributes["io.kubernetes.container.name"],
				event.Actor.Attributes["image"],
			)
		}
	}()
}

// containerOOMListener create channel to get events from containerd daemon.
// Listening is done within to goroutines. One for errors and second one for events.
// Channels will receive only oom events as defined in filters argument of Subscribe method.
// On oom event details are printed to stdout.
func containerdOOMListener(client *containerd.Client, wg *sync.WaitGroup) {
	ctx := context.Background()
	events, errs := client.Subscribe(ctx, "topic==/tasks/oom")
	go func() {
		defer wg.Done()
		for {
			cherr, ok := <-errs
			if !ok {
				log.Errorf("failed read containerd events errors channel")
				return
			}
			log.WithFields(log.Fields{"msg": "got containerd events errors channel event"}).Errorf("%+v", cherr)
		}
	}()
	go func() {
		defer wg.Done()
		for {
			event, ok := <-events
			if !ok {
				log.Errorf("failed read containerd events channel")
				return
			}
			fmt.Printf("%+v", event)
		}
	}()
	//TODO: check what data is in containerd oom event and if following code is needed.
	oom := namespaces.WithNamespace(context.Background(), "oom")
	allContainers, err := client.Containers(oom)
	if err != nil {
		log.WithFields(log.Fields{"msg": "failed get containers"}).Fatalf("error: %s", err)
	}
	for container := range allContainers {
		fmt.Fprintf(os.Stdout, "%+v", container)
	}
}
func main() {
	// wait group to allow goroutines listen on channels
	var wg sync.WaitGroup
	// check docker daemon socket exists
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		// create docker client with unix socket
		client, err := dockerclient.NewClient("unix:///var/run/docker.sock")
		if err != nil {
			log.WithFields(log.Fields{"msg": "failed create docker client"}).Fatalf("error: %s", err)
		}
		wg.Add(1)
		// listen for oom events
		dockerOOMListener(client, &wg)
		// if docker socket doesn't exists try attach to containerd socket
	} else if os.IsNotExist(err) {
		// check if containerd socket exists
		if _, err := os.Stat("/run/containerd/containerd.sock"); err == nil {
			// create containerd client with unix socket
			client, err := containerd.New("/run/containerd/containerd.sock")
			if err != nil {
				log.WithFields(log.Fields{"msg": "failed create containerd client"}).Fatalf("error: %s", err)
			}
			defer client.Close()
			wg.Add(2)
			// listen for oom events
			containerdOOMListener(client, &wg)
		} else {
			log.WithFields(log.Fields{"msg": "failed found container runtime socket"}).Errorf("%+v", err)
		}
	} else {
		log.WithFields(log.Fields{"msg": "failed found container runtime socket"}).Errorf("%+v", err)
	}
	wg.Wait()
}
