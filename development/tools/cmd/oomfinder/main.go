package main

import (
	"context"
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
	flag "github.com/spf13/pflag"

	// Register grpc event types
	_ "github.com/containerd/containerd/api/events"
)

// ContainerInfo is a subset of containerd.containers.Container fields. It's created from the containerd API call response.
// ContainerInfo holds minimal information needed to identify container and pod which had an event.
type ContainerInfo struct {
	ID        string `json:"containerId"`
	Image     string `json:"image"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Name      string `json:"name"`
}

// EventInfo holds minimal info about received event along with details about impacted container.
// It's created from received event and ContainerInfo object for respective container
type EventInfo struct {
	ContainerInfo ContainerInfo
	Message       string    `json:"message"`
	Timestamp     time.Time `json:"timestamp"`
}

var debug = flag.BoolP("debug", "d", false, "enable debug output")

// dockerOOMListener listen for events from docker daemon.
// Listening is done in a goroutine.
// Channel will receive only oom events as defined in Filters property of EventsOptions.
// On oom event, details are printed to stdout.
func dockerOOMListener(client *dockerclient.Client, wg *sync.WaitGroup) {
	eventsChannel := make(chan *dockerclient.APIEvents)
	// Filter events to get only OOM events for containers.
	eventOptions := dockerclient.EventsOptions{
		Filters: map[string][]string{
			"type":  {"container"},
			"event": {"oom"},
		},
	}
	// Subscribe to listen for docker deamon events.
	err := client.AddEventListenerWithOptions(eventOptions, eventsChannel)
	if err != nil {
		log.WithError(err).Error("got docker daemon events error")
	}
	log.Debug("Subscribed on docker socket")
	go func() {
		defer wg.Done()
		for {
			event, ok := <-eventsChannel
			if !ok {
				log.Error("failed read docker event")
				return
			}
			log.Debugf("Got docker oom event: %v", event)
			_, err := fmt.Fprintf(os.Stdout, "OOM event received, time: %s, namespace: %s, pod: %s, container: %s, image: %s\n",
				time.Unix(event.Time, 0).Format(time.RFC822Z),
				event.Actor.Attributes["io.kubernetes.pod.namespace"],
				event.Actor.Attributes["io.kubernetes.pod.name"],
				event.Actor.Attributes["io.kubernetes.container.name"],
				event.Actor.Attributes["image"],
			)
			if err != nil {
				log.WithError(err).Error("cannot print event details")
				continue
			}
		}
	}()
}

// containerOOMListener listen for events from containerd daemon.
// Listening is done within goroutine.
// Channels will receive only oom events as defined in filters argument of Subscribe method.
// On oom event details, are printed to stdout.
func containerdOOMListener(client *containerd.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	containers := make(map[string]ContainerInfo)
	ctx := context.Background()
	eventsClient := client.EventService()
	oomEventsCh, oomErrCh := eventsClient.Subscribe(ctx, "topic~=|^/tasks/oom|")
	log.Debug("Subscribed on containerd socket.")
	go func() {
		for {
			var e *events.Envelope
			select {
			case e = <-oomEventsCh:
			case err := <-oomErrCh:
				log.WithError(err).Error("got containerd event on errors channel")
				return
			}
			if e != nil {
				if e.Event != nil {
					log.Debug("Got containerd oom event.")
					v, err := typeurl.UnmarshalAny(e.Event)
					if err != nil {
						log.WithError(err).Warn("cannot unmarshal an event from Any")
						continue
					}
					log.Debugf("got oom event for containerID: %s\n", v.(*events2.TaskOOM).ContainerID)
					ctxWithNamespace := namespaces.WithNamespace(ctx, "k8s.io")
					container, err := client.ContainerService().Get(ctxWithNamespace, v.(*events2.TaskOOM).ContainerID)
					if err != nil {
						log.WithError(err).Error("Failed read container details.")
						continue
					}
					log.Debugf("Image: %s, Name: %s, Name: %s, Namespace: %s", container.Image, container.Labels["io.kubernetes.container.name"], container.Labels["io.kubernetes.pod.name"], container.Labels["io.kubernetes.pod.namespace"])
					containers[v.(*events2.TaskOOM).ContainerID] = ContainerInfo{
						ID:        v.(*events2.TaskOOM).ContainerID,
						Image:     container.Image,
						Namespace: container.Labels["io.kubernetes.pod.namespace"],
						Pod:       container.Labels["io.kubernetes.pod.name"],
						Name:      container.Labels["io.kubernetes.container.name"],
					}
					eventInfo := EventInfo{
						ContainerInfo: containers[v.(*events2.TaskOOM).ContainerID],
						Message:       "OOM event received",
						Timestamp:     e.Timestamp,
					}
					_, err = fmt.Fprintf(os.Stdout, "%s, time: %s, namespace: %s, pod: %s, container: %s, image: %s\n",
						eventInfo.Message,
						eventInfo.Timestamp,
						eventInfo.ContainerInfo.Namespace,
						eventInfo.ContainerInfo.Pod,
						eventInfo.ContainerInfo.Name,
						eventInfo.ContainerInfo.Image,
					)
					if err != nil {
						log.WithError(err).Error("cannot print EventInfo")
						continue
					}
				}
			}
		}
	}()
}

func main() {
	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	// Wait group to allow goroutines listen on channels.
	var wg sync.WaitGroup
	var socketFlag atomic.Value
	// Check docker daemon socket exists.
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		socketFlag.Store(1)
		// Create docker client with unix socket.
		client, err := dockerclient.NewClient("unix:///var/run/docker.sock")
		if err != nil {
			log.WithError(err).Fatalf("failed create docker client")
		}
		wg.Add(1)
		// Listen for oom events.
		dockerOOMListener(client, &wg)
	}
	// Check containerd daemon socket exists.
	if _, err := os.Stat("/run/containerd/containerd.sock"); err == nil {
		socketFlag.Store(1)
		// Create containerd client with unix socket.
		client, err := containerd.New("/run/containerd/containerd.sock")
		if err != nil {
			log.WithError(err).Fatalf("failed create containerd client")
		}
		wg.Add(1)
		// Listen for oom events.
		containerdOOMListener(client, &wg)
	}
	// Check if any container runtime socket was found
	if socketFlag.Load() == nil {
		log.Error("failed found container runtime socket")
	}
	wg.Wait()
}
