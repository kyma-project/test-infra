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
	"regexp"
	"strings"
)

// This message will be send by pubsub system.
type PubSubMessage struct {
	Message      MessagePayload `json:"message"`
	Subscription string         `json:"subscription"`
}

// This is the Message payload of pubsub message
type MessagePayload struct {
	Attributes   map[string]string `json:"attributes"`
	Data         string            `json:"data"` // This property is base64 encoded
	MessageId    string            `json:"messageId"`
	Message_Id   string            `json:"message_id"`
	PublishTime  string            `json:"publishTime"`
	Publish_time string            `json:"publish_time"`
}

// This is the Data payload of pubsub message payload.
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
	Message  string `json:"message"`
	Severity string `json:"severity,omitempty"`
	// Trace will be the same for one function call, you can use it for filetering in logs
	Trace  string            `json:"logging.googleapis.com/trace,omitempty"`
	Labels map[string]string `json:"logging.googleapis.com/operation,omitempty"`
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

const (
	projectID            = "sap-kyma-prow"
	oomEventFoundTopicId = "oom-event-found"
)

var (
	client             *storage.Client
	gcsPathRegex       *regexp.Regexp
	oomEventRegex      *regexp.Regexp
	oomEventFoundTopic *pubsub.Topic
)

// init will create clients reused by function calls
func init() {
	var err error
	ctx := context.Background()
	gcsPathRegex = regexp.MustCompile(`^gs://(.+?)/(.*)$`)
	oomEventRegex = regexp.MustCompile(`System OOM encountered`)
	// GCS client to read log files from bucket
	client, err = storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		log.Fatal(err)
	}
	// pubsub client to publish messages to pubsub
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}
	// topic to publish messages when oom event was found
	oomEventFoundTopic = pubsubClient.Topic(oomEventFoundTopicId)
	// disable batch sending by forcing publishing message when one message is ready to send
	oomEventFoundTopic.PublishSettings.CountThreshold = 1
	log.SetFlags(0)
}

func Checkoomevent(w http.ResponseWriter, r *http.Request) {
	// set trace value to use it in logEntry
	var trace string
	traceHeader := r.Header.Get("X-Cloud-Trace-Context")
	traceParts := strings.Split(traceHeader, "/")
	if len(traceParts) > 0 && len(traceParts[0]) > 0 {
		trace = fmt.Sprintf("projects/%s/traces/%s", projectID, traceParts[0])
	}

	functionCtx := context.Background()
	// decode http messages body
	var message PubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		log.Println(LogEntry{
			Message:   "failed decode message body",
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.checkoomevent",
			Labels:    map[string]string{"messageId": message.Message.MessageId},
		})
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprint(w, "500 - failed decode message!"); err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
				Severity:  "CRITICAL",
				Trace:     trace,
				Labels:    map[string]string{"messageId": message.Message.MessageId},
				Component: "kyma.prow.cloud-function.checkoomevent",
			})
		}
		return
	}
	if message.Message.Data == "" {
		log.Println(LogEntry{
			Message:   "message data is empty, nothing to analyse",
			Severity:  "ERROR",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.checkoomevent",
			Labels:    map[string]string{"messageId": message.Message.MessageId},
		})
		w.WriteHeader(http.StatusBadRequest)
		if _, err := fmt.Fprint(w, "400 - message is empty!"); err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
				Severity:  "CRITICAL",
				Trace:     trace,
				Labels:    map[string]string{"messageId": message.Message.MessageId},
				Component: "kyma.prow.cloud-function.checkoomevent",
			})
		}
		return
	}

	// got valid message
	log.Println(LogEntry{
		Message:   fmt.Sprintf("received message with id: %s", message.Message.MessageId),
		Severity:  "INFO",
		Trace:     trace,
		Labels:    map[string]string{"messageId": message.Message.MessageId},
		Component: "kyma.prow.cloud-function.checkoomevent",
	})

	// decode base64 prow message
	bdata, err := base64.StdEncoding.DecodeString(message.Message.Data)
	if err != nil {
		log.Println(LogEntry{
			Message:   "prow message data field base64 decoding failed",
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.checkoomevent",
			Labels:    map[string]string{"messageId": message.Message.MessageId},
		})
		return
	}

	// get message payload published by prow
	var data ProwMessage
	if err := json.Unmarshal(bdata, &data); err != nil {
		log.Println(LogEntry{
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.checkoomevent",
			Message:   "failed unmarshal message data to json",
			Trace:     trace,
			Labels:    map[string]string{"messageId": message.Message.MessageId},
		})
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := fmt.Fprint(w, "500 - failed unmarshal message data to json!"); err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
				Severity:  "CRITICAL",
				Trace:     trace,
				Labels:    map[string]string{"messageId": message.Message.MessageId},
				Component: "kyma.prow.cloud-function.checkoomevent",
			})
		}
		return
	}

	// check prowjob status to know if oom event search can start
	if data.Status != "success" && data.Status != "failure" {
		// prowjob didn't finish no data to search for oom
		log.Println(LogEntry{
			Severity:  "INFO",
			Component: "kyma.prow.cloud-function.checkoomevent",
			Message:   fmt.Sprintf("prowjob %s status is %s, no data to search, skipping", data.JobName, data.Status),
			Trace:     trace,
			Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
		})
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, "200 - message processed, prowjob not finished, no data to analyse"); err != nil {
			log.Println(LogEntry{
				Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
				Severity:  "CRITICAL",
				Trace:     trace,
				Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
				Component: "kyma.prow.cloud-function.checkoomevent",
			})
		}
		return
	} else {
		// start looking for oom event
		log.Println(LogEntry{
			Severity:  "INFO",
			Component: "kyma.prow.cloud-function.checkoomevent",
			Message:   fmt.Sprintf("prowjob %s status is %s, searching oom events", data.JobName, data.Status),
			Trace:     trace,
			Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
		})
		// extract path from gcs bucket url
		gcsMatch := gcsPathRegex.FindStringSubmatch(data.GcsPath)
		// build gcs bucket path to describe nodes output
		objectPath := fmt.Sprintf("%s/artifacts/describe_nodes.txt", gcsMatch[2])
		// get object with describe nodes output from bucket
		rc, err := client.Bucket(gcsMatch[1]).Object(objectPath).NewReader(functionCtx)
		// check if object in a bucket exist
		if err == storage.ErrObjectNotExist {
			log.Println(LogEntry{
				Severity:  "INFO",
				Component: "kyma.prow.cloud-function.checkoomevent",
				Message:   "describe_nodes.txt not found, no data to analyse",
				Trace:     trace,
				Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
			})
			w.WriteHeader(http.StatusOK)
			if _, err := fmt.Fprintf(w, "200 - message processed, describe_nodes.txt not found, no data to analyse"); err != nil {
				log.Println(LogEntry{
					Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
					Severity:  "CRITICAL",
					Trace:     trace,
					Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
					Component: "kyma.prow.cloud-function.checkoomevent",
				})
			}
			return
		} else if err != nil && err != storage.ErrObjectNotExist {
			// report other errors than object not exist errors
			log.Println(LogEntry{
				Severity:  "CRITICAL",
				Component: "kyma.prow.cloud-function.checkoomevent",
				Message:   fmt.Sprintf("failed get describe_nodes.txt from gcs, can't analyse data, error: %s", err.Error()),
				Trace:     trace,
				Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
			})
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprintf(w, "500 - failed get describe_nodes.txt from gcs, can't analyse data"); err != nil {
				log.Println(LogEntry{
					Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
					Severity:  "CRITICAL",
					Trace:     trace,
					Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
					Component: "kyma.prow.cloud-function.checkoomevent",
				})
			}
			return
		} else {
			defer rc.Close()
			// read content of descrbie_nodes.txt
			body, err := ioutil.ReadAll(rc)
			if err != nil {
				log.Println(LogEntry{
					Severity:  "CRITICAL",
					Component: "kyma.prow.cloud-function.checkoomevent",
					Message:   fmt.Sprintf("failed read describe_nodes.txt, can't analyse data, error: %s", err.Error()),
					Trace:     trace,
					Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
				})
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := fmt.Fprintf(w, "500 - failed read describe_nodes.txt from gcs, can't analyse data"); err != nil {
					log.Println(LogEntry{
						Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
						Severity:  "CRITICAL",
						Trace:     trace,
						Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
						Component: "kyma.prow.cloud-function.checkoomevent",
					})
				}
				return
			}

			// search for oom event in describe_nodes.txt
			oomFound := oomEventRegex.Match(body)
			if oomFound {
				// oom event found
				log.Println(LogEntry{
					Severity:  "INFO",
					Component: "kyma.prow.cloud-function.checkoomevent",
					Message:   fmt.Sprintf("oom event found in prowjob %s", data.JobName),
					Trace:     trace,
					Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
				})
				// publish message to oom-event-found topic
				pubsubResponse := oomEventFoundTopic.Publish(functionCtx, &pubsub.Message{
					Data: bdata,
				})
				// check if message published successfully
				msgID, err := pubsubResponse.Get(functionCtx)
				if err != nil {
					log.Println(LogEntry{
						Severity:  "CRITICAL",
						Component: "kyma.prow.cloud-function.checkoomevent",
						Message:   fmt.Sprintf("failed publish oom event found message to oom-event-found topic, error: %s", err.Error()),
						Trace:     trace,
						Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
					})
					w.WriteHeader(http.StatusInternalServerError)
					if _, err := fmt.Fprintf(w, "500 - failed publish message to oom-event-found topic"); err != nil {
						log.Println(LogEntry{
							Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
							Severity:  "CRITICAL",
							Trace:     trace,
							Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
							Component: "kyma.prow.cloud-function.checkoomevent",
						})
					}
					return
				}
				// log and Ack pubsub message
				log.Println(LogEntry{
					Severity:  "INFO",
					Component: "kyma.prow.cloud-function.checkoomevent",
					Message:   fmt.Sprintf("published oom event found message to oom-event-found topic, messageId: %s", msgID),
					Trace:     trace,
					Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
				})
				w.WriteHeader(http.StatusOK)
				if _, err := fmt.Fprintf(w, "200 - message processed, published message to oom-event-found topic"); err != nil {
					log.Println(LogEntry{
						Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
						Severity:  "CRITICAL",
						Trace:     trace,
						Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
						Component: "kyma.prow.cloud-function.checkoomevent",
					})
				}
				return
			} else {
				// oom event was not found
				// log and Ack pubsub message
				log.Println(LogEntry{
					Severity:  "INFO",
					Component: "kyma.prow.cloud-function.checkoomevent",
					Message:   fmt.Sprintf("oom event not found in prowjob %s", data.JobName),
					Trace:     trace,
					Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
				})
				w.WriteHeader(http.StatusOK)
				if _, err := fmt.Fprintf(w, "200 - message processed"); err != nil {
					log.Println(LogEntry{
						Message:   fmt.Sprintf("failed send response for message id: %s", message.Message.MessageId),
						Severity:  "CRITICAL",
						Trace:     trace,
						Labels:    map[string]string{"runID": data.RunID, "prowjobName": data.JobName, "messageId": message.Message.MessageId},
						Component: "kyma.prow.cloud-function.checkoomevent",
					})
				}
				return
			}
		}
	}
}
