package getGithubIssue

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v36/github"
	"github.com/kyma-project/test-infra/development/gcp/pkg/cloudfunctions"
	kymapubsub "github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"golang.org/x/oauth2"
	"os"
)

var (
	firestoreClient         *firestore.Client
	pubSubClient            *pubsub.Client
	githubClient            *github.Client
	ts                      oauth2.TokenSource
	projectID               string
	githubAccessToken       string
	githubOrg               string
	githubRepo              string
	firestoreCollection     string
	getGithubCommiterTopic  string
	getFailureInstanceTopic string
)

func init() {
	var err error
	projectID = os.Getenv("GCP_PROJECT_ID")
	githubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	githubOrg = os.Getenv("GITHUB_ORG")
	githubRepo = os.Getenv("GITHUB_REPO")
	firestoreCollection = os.Getenv("FIRESTORE_COLLECTION")
	getGithubCommiterTopic = os.Getenv("GET_GITHUB_COMMITER_TOPIC")
	getFailureInstanceTopic = os.Getenv("GET_FAILURE_INSTANCE_TOPIC")
	ctx := context.Background()
	if getGithubCommiterTopic == "" {
		panic("environment variable GET_GITHUB_COMMITER_TOPIC is empty")
	}
	if getFailureInstanceTopic == "" {
		panic("environment variable GET_FAILURE_INSTANCE_TOPIC is empty")
	}
	if projectID == "" {
		panic("environment variable GCP_PROJECT_ID is empty")
	}
	if githubAccessToken == "" {
		panic("environment variable GITHUB_ACCESS_TOKEN is empty")
	}
	if githubOrg == "" {
		panic("environment variable GITHUB_ORG is empty")
	}
	if githubRepo == "" {
		panic("environment variable GITHUB_REPO is empty")
	}
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed creating firestore client, error: %s", err.Error()))
	}
	pubSubClient, err = pubsub.NewClient(ctx, projectID)
	if err != nil {
		panic(fmt.Sprintf("Failed creating pubsub client, error: %s", err.Error()))
	}
	ts = oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient = github.NewClient(tc)
}

// TODO: this function must be rewritten
func isGithubIssueOpen(ctx context.Context, client *github.Client, message kymapubsub.FailingTestMessage, githubOrg, githubRepo string) (*bool, *github.Issue, error) {
	ghIssue, ghResponse, err := client.Issues.Get(ctx, githubOrg, githubRepo, int(*message.GithubIssueNumber))
	if ghResponse != nil {
		err = github.CheckResponse(ghResponse.Response)
		if err != nil {
			return nil, nil, fmt.Errorf("github API call reply with error, error: %w", err)
		}
	} else if err != nil {
		return nil, nil, fmt.Errorf("calling github API failed, error: %w", err)
	}
	b := new(bool)
	if *ghIssue.State == "open" {
		b = github.Bool(true)
	} else {
		b = github.Bool(false)
	}
	return b, ghIssue, nil
}

func createGithubIssue(ctx context.Context, client *github.Client, message kymapubsub.FailingTestMessage, githubOrg, githubRepo string) (*github.Issue, error) {
	issueRequest := &github.IssueRequest{
		Title:     github.String(fmt.Sprintf("Failed prowjob: %s", *message.JobName)),
		Body:      nil,
		Labels:    &[]string{"test-failing", "ci-force-bot"},
		Assignee:  nil,
		State:     github.String("open"),
		Milestone: nil,
		Assignees: nil,
	}
	issue, ghResponse, err := client.Issues.Create(ctx, githubOrg, githubRepo, issueRequest)
	if ghResponse != nil {
		err = github.CheckResponse(ghResponse.Response)
		if err != nil {
			return nil, fmt.Errorf("github API call reply with error, error: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("callilg github API failed, error: %w", err)
	}
	return issue, nil
}

func GetGithubIssue(ctx context.Context, m kymapubsub.MessagePayload) error {
	var err error
	// set trace value to use it in logEntry
	var failingTestMessage kymapubsub.FailingTestMessage
	logger := cloudfunctions.NewLogger()
	logger.WithComponent("kyma.prow.cloud-function.Getfailureinstancedetails")
	logger.GenerateTraceValue(projectID, "GetGithubIssue")

	contextMetadata, err := metadata.FromContext(ctx)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed extract metadata from function call context, error: %s", err.Error()))
	}
	logger.WithLabel("messageId", contextMetadata.EventID)
	// Decode
	err = json.Unmarshal(m.Data, &failingTestMessage)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed unmarshal message data field, error: %s", err.Error()))
	}
	logger.WithLabel("prowjobName", *failingTestMessage.JobName)

	jobID, err := kymapubsub.GetJobId(failingTestMessage.URL)
	if err != nil {
		logger.LogCritical(fmt.Sprintf("failed get job ID, error: %s", err.Error()))
	}
	logger.WithLabel("jobID", *jobID)
	logger.LogInfo(fmt.Sprintf("found prowjob execution ID: %s", *jobID))

	// sprawdz czy message ma gh issue
	if failingTestMessage.GithubIssueNumber != nil {
		// jeśli message ma gh issue sprawdź czy otwarte
		open, ghIssue, err := isGithubIssueOpen(ctx, githubClient, failingTestMessage, githubOrg, githubRepo)
		if err != nil {
			logger.LogError(fmt.Sprintf("failed get github issue status, error: %s", err.Error()))
		}
		logger.LogInfo(fmt.Sprintf("github issue number %d has status: %s", ghIssue.GetNumber(), ghIssue.GetState()))
		if !*open {
			// stary failure instance w firestore oznacz jako zamknięty
			docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", firestoreCollection, *failingTestMessage.FirestoreDocumentID))
			_, err = docRef.Set(ctx, map[string]bool{"open": false}, firestore.Merge([]string{"open"}))
			// usuń referencje do starego dokumentu w firestore
			failingTestMessage.FirestoreDocumentID = nil
			// jeśli zamknięte to utwórz i dodaj do firestore
			ghIssue, err := createGithubIssue(ctx, githubClient, failingTestMessage, githubOrg, githubRepo)
			if err != nil {
				logger.LogError(fmt.Sprintf("failed create github issue, error: %s", err.Error()))
			}
			if ghIssue != nil {
				logger.LogInfo(fmt.Sprintf("github issue created. issue number: %d", ghIssue.GetNumber()))
				if ghIssue.Number != nil {
					failingTestMessage.GithubIssueNumber = github.Int64(int64(*ghIssue.Number))
				}
			}
			// opublikuj wiadomość do topicu getfailureinstance
			publlishedMessageID, err := kymapubsub.PublishPubSubMessage(ctx, pubSubClient, failingTestMessage, getFailureInstanceTopic)
			if err != nil {
				// log error publishing message to pubsub
				logger.LogCritical(fmt.Sprintf("failed publishiing to pubsub, error: %s", err.Error()))
			}
			// log publishing message to pubsub
			logger.LogInfo(fmt.Sprintf("published pubsub message to topic %s, id: %s", getFailureInstanceTopic, *publlishedMessageID))
		}
	} else {
		// jeśli message nie ma gh issue to utwórz i dodaj do firestore
		ghIssue, err := createGithubIssue(ctx, githubClient, failingTestMessage, githubOrg, githubRepo)
		if err != nil {
			logger.LogError(fmt.Sprintf("failed create github issue, error: %s", err.Error()))
		}
		if ghIssue != nil {
			logger.LogInfo(fmt.Sprintf("github issue created. issue number: %d", ghIssue.GetNumber()))
			if ghIssue.Number != nil {
				failingTestMessage.GithubIssueNumber = github.Int64(int64(*ghIssue.Number))
				docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", firestoreCollection, *failingTestMessage.FirestoreDocumentID))
				_, err = docRef.Set(ctx, map[string]int{"githubIssueNumber": *ghIssue.Number}, firestore.Merge([]string{"githubIssueNumber"}))
				if err != nil {
					// log error
					logger.LogError(fmt.Sprintf("failed adding github issue number %d, to failing test instance, error: %s", ghIssue.GetNumber(), err.Error()))
					// TODO: need error reporting for such case, without failing whole function
				} else {
					// log gh issue was added to firestore
					logger.LogError(fmt.Sprintf("github issue, number %d, added to failing test instance", ghIssue.GetNumber()))
				}
			} else {
				// log error ghIssue.Number is nil
				logger.LogError(fmt.Sprintf("github issue number is nil, something went wrong with creating github issue"))
				// TODO: need error reporting for such case, without failing whole function
			}
		} else {
			// log error getting ghIssue after creating
			logger.LogError(fmt.Sprintf("github issue is nil, something went wrong with creating it"))
			// TODO: need error reporting for such case, without failing whole function
		}
		publlishedMessageID, err := kymapubsub.PublishPubSubMessage(ctx, pubSubClient, failingTestMessage, getGithubCommiterTopic)
		if err != nil {
			// log error publishing message to pubsub
			logger.LogCritical(fmt.Sprintf("failed publishing to pubsub, error: %s", err.Error()))
		}
		// log publishin message to pubsub
		logger.LogInfo(fmt.Sprintf("published pubsub message to topic %s, id: %s", getGithubCommiterTopic, *publlishedMessageID))
	}

	//znajdz commitera - zrób to w osobnej funkcji

	//dodaj do gh issue koemntarz z linkiem do kolejnego wystąpienia błędu, - zrób to w osobnej funkcji
	//link do url
	//nazwa testu
	//czas uruchomienia
	//rodzaj testu
	//base sha
	//pr number
	//commiter
	//jakie logi można dodać w komentarzu. Może oprzeć się na junit lens z prow?
	return nil
}
