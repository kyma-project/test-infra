package checkoomevent

import (
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type PubSubMessage struct {
	Message      MessagePayload `json:"message"`
	Subscription string         `json:"subscription"`
}

type MessagePayload struct {
	Attributes   map[string]string `json:"attributes"`
	Data         string            `json:"data"`
	MessageId    string            `json:"messageId"`
	Message_Id   string            `json:"message_id"`
	PublishTime  string            `json:"publishTime"`
	Publish_time string            `json:"publish_time"`
}

type ProwMessage struct {
	Project string                   `json:"project"`
	Topic   string                   `json:"topic"`
	RunID   string                   `json:"runid"`
	Status  string                   `json:"status"`
	URL     string                   `json:"url"`
	GcsPath string                   `json:"gcs_path"`
	Refs    []map[string]interface{} `json:"refs"`
	JobType string                   `json:"job_type"`
	JobName string                   `json:"job_name"`
}

// Entry defines a log entry.
type LogEntry struct {
	Message  string            `json:"message"`
	Severity string            `json:"severity,omitempty"`
	Trace    string            `json:"logging.googleapis.com/trace,omitempty"`
	Labels   map[string]string `json:"logging.googleapis.com/operation,omitempty"`
	// Cloud Log Viewer allows filtering and display of this as `jsonPayload.component`.
	Component string `json:"component,omitempty"`
}

// String renders an entry structure to the JSON format expected by Cloud Logging.
func (e LogEntry) String() string {
	if e.Severity == "" {
		e.Severity = "INFO"
	}
	out, err := json.Marshal(e)
	if err != nil {
		log.Printf("json.Marshal: %v", err)
	}
	return string(out)
}

var (
	client             *storage.Client
	gcsPathRegex       *regexp.Regexp
	oomEventRegex      *regexp.Regexp
	message            PubSubMessage
	data               ProwMessage
	oomEventFoundTopic *pubsub.Topic
	pubsubResults      []*pubsub.PublishResult
	projectID          string
)

func init() {
	var err error
	ctx := context.Background()
	gcsPathRegex = regexp.MustCompile(`^gs://(.+?)/(.*)$`)
	oomEventRegex = regexp.MustCompile(`System OOM encountered`)
	client, err = storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		log.Fatal(err)
	}
	pubsubClient, err := pubsub.NewClient(ctx, "sap-kyma-prow")
	if err != nil {
		log.Fatal(err)
	}
	oomEventFoundTopic = pubsubClient.Topic("oom-event-found")
	log.SetFlags(0)
	projectID = "sap-kyma-prow"
}

func Checkoomevent(w http.ResponseWriter, r *http.Request) {
	var trace string
	if projectID != "" {
		traceHeader := r.Header.Get("X-Cloud-Trace-Context")
		traceParts := strings.Split(traceHeader, "/")
		if len(traceParts) > 0 && len(traceParts[0]) > 0 {
			trace = fmt.Sprintf("projects/%s/traces/%s", projectID, traceParts[0])
		}
	}
	var message PubSubMessage
	functionCtx := context.Background()
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "500 - failed decode message!")
		return
	}
	if message.Message.Data == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "400 - message is empty!")
		return
	}
	bdata, err := base64.StdEncoding.DecodeString(message.Message.Data)
	if err != nil {
		log.Println(LogEntry{
			Message:   "prow message data field base64 decoding failed",
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.checkoomevent",
			Labels:    map[string]string{"messageId": message.Message.MessageId},
		})
	}
	if err := json.Unmarshal(bdata, &data); err != nil {
		fmt.Fprintf(os.Stdout, "json unmarshal failed")
		log.Fatal(err)
	}
	if data.Status != "success" && data.Status != "failure" {
		// prowjob didn't finish no data to search for oom
		log.Println(LogEntry{
			Severity:  "INFO",
			Component: "kyma.prow.cloud-function.checkoomevent",
			Message:   fmt.Sprintf("prowjob %s status is %s, no data to search, skipping", data.JobName, data.Status),
			Trace:     trace,
			Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName},
		})
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "200 - message processed, prowjob not finished, no data to analyse")
	} else {
		log.Println(LogEntry{
			Severity:  "INFO",
			Component: "kyma.prow.cloud-function.checkoomevent",
			Message:   fmt.Sprintf("prowjob %s status is %s, searching oom events", data.JobName, data.Status),
			Trace:     trace,
			Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName},
		})
		// extract path from gcs bucket url
		gcsMatch := gcsPathRegex.FindStringSubmatch(data.GcsPath)
		// build gcs bucket path to describe nodes output
		objectPath := fmt.Sprintf("%s/artifacts/describe_nodes.txt", gcsMatch[2])
		// get object with describe nodes output from bucket
		rc, err := client.Bucket(gcsMatch[1]).Object(objectPath).NewReader(functionCtx)
		defer rc.Close()
		// check if object exist
		if err == storage.ErrObjectNotExist {
			log.Println(LogEntry{
				Severity:  "INFO",
				Component: "kyma.prow.cloud-function.checkoomevent",
				Message:   "describe_nodes.txt not found, no data to analyse",
				Trace:     trace,
				Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName},
			})
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "200 - message processed, describe_nodes.txt not found, no data to analyse")
		}
		if err != nil && err != storage.ErrObjectNotExist {
			log.Println(LogEntry{
				Severity:  "CRITICAL",
				Component: "kyma.prow.cloud-function.checkoomevent",
				Message:   fmt.Sprintf("failed get describe_nodes.txt from gcs, can't analyse data, error: %s", err.Error()),
				Trace:     trace,
				Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName},
			})
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 - failed get describe_nodes.txt from gcs, can't analyse data")
		}

		// read content of descrbie_nodes.txt
		body, err := ioutil.ReadAll(rc)
		if err != nil {
			log.Println(LogEntry{
				Severity:  "CRITICAL",
				Component: "kyma.prow.cloud-function.checkoomevent",
				Message:   fmt.Sprintf("failed read describe_nodes.txt, can't analyse data, error: %s", err.Error()),
				Trace:     trace,
				Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName},
			})
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 - failed read describe_nodes.txt from gcs, can't analyse data")
		}

		// read for oom event in describe_nodes.txt
		oomFound := oomEventRegex.Match(body)
		if oomFound {
			log.Println(LogEntry{
				Severity:  "INFO",
				Component: "kyma.prow.cloud-function.checkoomevent",
				Message:   fmt.Sprintf("oom event found in prowjob %s", data.JobName),
				Trace:     trace,
				Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName},
			})
			defer oomEventFoundTopic.Stop()
			pubsubResponse := oomEventFoundTopic.Publish(functionCtx, &pubsub.Message{
				Data: bdata,
			})
			pubsubResults = append(pubsubResults, pubsubResponse)
			for _, result := range pubsubResults {
				msgID, err := result.Get(functionCtx)
				if err != nil {
					log.Println(LogEntry{
						Severity:  "CRITICAL",
						Component: "kyma.prow.cloud-function.checkoomevent",
						Message:   fmt.Sprintf("failed publish oom event found message to oom-event-found topic, error: %s", err.Error()),
						Trace:     trace,
						Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName},
					})
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "500 - failed publish message to oom-event-found topic")
				}
				log.Println(LogEntry{
					Severity:  "INFO",
					Component: "kyma.prow.cloud-function.checkoomevent",
					Message:   fmt.Sprintf("published oom event found message to oom-event-found topic, messageId: %s", msgID),
					Trace:     trace,
					Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName},
				})
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "200 - message processed, published message to oom-event-found topic")
			}
		} else {
			log.Println(LogEntry{
				Severity:  "INFO",
				Component: "kyma.prow.cloud-function.checkoomevent",
				Message:   fmt.Sprintf("oom event not found in prowjob %s", data.JobName),
				Trace:     trace,
				Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName},
			})
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "200 - message processed")
		}
	}
}
