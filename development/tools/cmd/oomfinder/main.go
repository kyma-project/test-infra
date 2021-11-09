package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/containerd/containerd"
	events2 "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/typeurl"
	dockerclient "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"

	// Register grpc event types
	_ "github.com/containerd/containerd/api/events"
)

// TODO: Update documentation for functions

// TODO: Document new types
type ContainerInfo struct {
	ID        string `json:"containerId"`
	Image     string `json:"image"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Name      string `json:"name"`
}

type EventInfo struct {
	ContainerInfo ContainerInfo
	Message       string    `json:"message"`
	Timestamp     time.Time `json:"timestamp"`
}

// dockerOOMListener create channel to get events from docker daemon.
// Listening is done in a goroutine.
// Channel will receive only oom events as defined in Filters property of EventsOptions.
// On oom event details are printed to stdout.
func dockerOOMListener(client *dockerclient.Client, wg *sync.WaitGroup) {
	// TODO: Align output with containerdOOMListener
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
	defer wg.Done()
	containers := make(map[string]ContainerInfo)
	ctx := context.Background()
	eventsClient := client.EventService()
	oomEventsCh, oomErrCh := eventsClient.Subscribe(ctx, "topic~=|^/tasks/oom|")
	createEventsCh, createErrCh := eventsClient.Subscribe(ctx, "topic~=|^/tasks/create|")
	log.Debug("Subscribed on containerd socket.")
	go func() {
		for {
			var e *events.Envelope
			// var v interface{}
			select {
			case e = <-oomEventsCh:
			case err := <-oomErrCh:
				log.WithError(err).Error("got containerd event on errors channel")
				return
			}
			if e != nil {
				var out []byte
				if e.Event != nil {
					log.Debug("Got containerd oom event.")
					v, err := typeurl.UnmarshalAny(e.Event)
					if err != nil {
						log.WithError(err).Warn("cannot unmarshal an event from Any")
						continue
					}
					log.Debugf("got oom event for containerID: %s\n", v.(*events2.TaskOOM).ContainerID)
					eventInfo := EventInfo{
						ContainerInfo: containers[v.(*events2.TaskOOM).ContainerID],
						Message:       "OOM event received",
						Timestamp:     e.Timestamp,
					}
					out, err = json.Marshal(eventInfo)
					if err != nil {
						log.WithError(err).Error("cannot marshal EventInfo into JSON")
						continue
					}
				}
				if _, err := fmt.Println(string(out)); err != nil {
					log.WithError(err).Error("Failed print EventInfo content.")
					return
				}
			}
		}
	}()

	go func() {
		for {
			var e *events.Envelope
			// var v interface{}
			select {
			case e = <-createEventsCh:
			case err := <-createErrCh:
				log.WithError(err).Error("got containerd event on errors channel")
				return
			}
			if e != nil {
				if e.Event != nil {
					log.Debug("Got containerd task create event.")
					v, err := typeurl.UnmarshalAny(e.Event)
					if err != nil {
						log.WithError(err).Warn("cannot unmarshal an event from Any")
						continue
					}
					log.Debugf("got task create event for containerID: %s\n", v.(*events2.TaskCreate).ContainerID)
					ctxWithNamespace := namespaces.WithNamespace(ctx, "k8s.io")
					container, err := client.ContainerService().Get(ctxWithNamespace, v.(*events2.TaskCreate).ContainerID)
					if err != nil {
						log.WithError(err).Error("Failed read container details.")
						continue
					}
					log.Debugf("Image: %s, Name: %s, Name: %s, Namespace: %s", container.Image, container.Labels["io.kubernetes.container.name"], container.Labels["io.kubernetes.pod.name"], container.Labels["io.kubernetes.pod.namespace"])
					containers[v.(*events2.TaskCreate).ContainerID] = ContainerInfo{
						ID:        v.(*events2.TaskCreate).ContainerID,
						Image:     container.Image,
						Namespace: container.Labels["io.kubernetes.pod.namespace"],
						Pod:       container.Labels["io.kubernetes.pod.name"],
						Name:      container.Labels["io.kubernetes.container.name"],
					}
				}
			}
		}
	}()
}

func main() {
	// TODO: Add flag to enable debug logging
	// wait group to allow goroutines listen on channels
	var wg sync.WaitGroup
	var socketFlag atomic.Value
	// check docker daemon socket exists
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		socketFlag.Store(1)
		// create docker client with unix socket
		client, err := dockerclient.NewClient("unix:///var/run/docker.sock")
		if err != nil {
			log.WithError(err).Fatalf("failed create docker client")
		}
		wg.Add(1)
		// listen for oom events
		dockerOOMListener(client, &wg)
		// if docker socket doesn't exists try attach to containerd socket
	}
	if _, err := os.Stat("/run/containerd/containerd.sock"); err == nil {
		socketFlag.Store(1)
		// create containerd client with unix socket
		client, err := containerd.New("/run/containerd/containerd.sock")
		if err != nil {
			log.WithError(err).Fatalf("failed create containerd client")
		}
		wg.Add(1)
		// listen for oom events
		containerdOOMListener(client, &wg)
	}
	if socketFlag.Load() == nil {
		log.Error("failed found container runtime socket")
	}
	wg.Wait()
}
