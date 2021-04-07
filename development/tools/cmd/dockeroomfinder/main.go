package main

import (
	"fmt"
	dockerclient "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	client        *dockerclient.Client
	eventOptions  dockerclient.EventsOptions
	wg            sync.WaitGroup
	eventsChannel chan *dockerclient.APIEvents
)

func init() {
	var err error
	client, err = dockerclient.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.WithFields(log.Fields{"msg": "failed create docker client"}).Fatalf("error: %s", err)
	}
}

func main() {
	wg.Add(1)
	eventsChannel = make(chan *dockerclient.APIEvents)
	eventOptions = dockerclient.EventsOptions{
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
			fmt.Printf("OOM event received time: %s , namespace: %s , pod: %s ,container: %s , image: %s ",
				time.Unix(event.Time, 0).Format(time.RFC822Z),
				event.Actor.Attributes["io.kubernetes.pod.namespace"],
				event.Actor.Attributes["io.kubernetes.pod.name"],
				event.Actor.Attributes["io.kubernetes.container.name"],
				event.Actor.Attributes["image"],
			)
		}
	}()
	wg.Wait()
}
