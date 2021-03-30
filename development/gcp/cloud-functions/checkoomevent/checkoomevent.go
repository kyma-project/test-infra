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
	Project string              `json:"project"`
	Topic   string              `json:"topic"`
	RunID   string              `json:"runid"`
	Status  string              `json:"status"`
	URL     string              `json:"url"`
	GcsPath string              `json:"gcs_path"`
	Refs    []map[string]string `json:"refs"`
	JobType string              `json:"job_type"`
	JobName string              `json:"job_name"`
}

var (
	client             *storage.Client
	gcsPathRegex       *regexp.Regexp
	oomEventRegex      *regexp.Regexp
	message            PubSubMessage
	data               ProwMessage
	oomEventFoundTopic *pubsub.Topic
	res                *pubsub.PublishResult
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
}

func Checkoomevent(w http.ResponseWriter, r *http.Request) {
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
	fmt.Fprintf(os.Stdout, "%s", message.Message.Data)
	bdata, err := base64.StdEncoding.DecodeString(message.Message.Data)
	if err != nil {
		fmt.Fprintf(os.Stdout, "base64 decoding failed")
		log.Fatal(err)
	}
	if err := json.Unmarshal(bdata, &data); err != nil {
		fmt.Fprintf(os.Stdout, "json unmarshal failed")
		log.Fatal(err)
	}
	gcsMatch := gcsPathRegex.FindStringSubmatch(data.GcsPath)
	// Read the object1 from bucket.
	fmt.Fprintf(os.Stdout, "%v", data)
	fmt.Fprintf(os.Stdout, "%v", gcsMatch)
	objectPath := fmt.Sprintf("%s/artifacts/describe_nodes.txt", gcsMatch[2])
	rc, err := client.Bucket(gcsMatch[1]).Object(objectPath).NewReader(functionCtx)
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()
	body, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}
	oomFound := oomEventRegex.Match(body)
	if oomFound {
		fmt.Fprintf(os.Stdout, "OOM event detected")
		res = oomEventFoundTopic.Publish(functionCtx, &pubsub.Message{
			Data: bdata,
		})
	} else {
		fmt.Fprintf(os.Stdout, "OOM event not found")
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "500 - failed decode data field")
	}
	if res != nil {
		msgID, err := res.Get(functionCtx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "500 - failed publish oom-event-found topic")
			log.Fatal(err)
		}
		if msgID != "" {
			fmt.Fprintf(os.Stdout, "%s", msgID)
		}
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "200 - message processed")
}
