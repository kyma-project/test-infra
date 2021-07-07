package getGithubIssue

import "golang.org/x/oauth2"

var (
	firestoreClient   *firestore.Client
	githubClient      *github.Client
	projectID         string
	githubAccessToken string
	githubOrg         string
	githubRepo        string
)





func checkGithubIssueStatus(ctx context.Context, client *github.Client, message ProwMessage, githubOrg, githubRepo, trace, eventID, jobID string, githubIssueNumber interface{}) (*bool, error) {
	issue, response, err := client.Issues.Get(ctx, githubOrg, githubRepo, githubIssueNumber.(int))
	if err != nil {
		log.Println(response)
		log.Println(LogEntry{
			Message:   fmt.Sprintf("could not get github issue number %d, error: %s",githubIssueNumber.(int), err.Error()),
			Severity:  "CRITICAL",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": message.JobName},
		})
		return nil, err
	} else {
		log.Println(response)
		log.Println(LogEntry{
			Message:   fmt.Sprintf("github issue number %d has status: %s",githubIssueNumber.(int), *issue.State),
			Severity:  "INFO",
			Trace:     trace,
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
			Labels:    map[string]string{"messageId": eventID, "jobID": jobID, "prowjobName": message.JobName},
		})
		b := new(bool)
		if *issue.State == "open" {
			*b = true
			return b, nil
		}
		return b, nil
	}
}

func init() {
	var err error
	projectID = os.Getenv("GCP_PROJECT_ID")
	githubAccessToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	githubOrg = os.Getenv("GITHUB_ORG")
	githubRepo = os.Getenv("GITHUB_REPO")
	ctx := context.Background()
	if projectID == "" {
		log.Println(LogEntry{
			Message:   "environment variable GCP_PROJECT_ID is empty, can't setup firebase client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GCP_PROJECT_ID is empty, can't setup firebase client")
	}
	if githubAccessToken == "" {
		log.Println(LogEntry{
			Message:   "environment variable GITHUB_ACCESS_TOKEN is empty, can't setup github client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GITHUB_ACCESS_TOKEN is empty, can't setup github client")
	}
	if githubOrg == "" {
		log.Println(LogEntry{
			Message:   "environment variable GITHUB_ORG is empty, can't setup github client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GITHUB_ACCESS_TOKEN is empty, can't setup github client")
	}
	if githubRepo == "" {
		log.Println(LogEntry{
			Message:   "environment variable GITHUB_REPO is empty, can't setup github client",
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic("environment variable GITHUB_ACCESS_TOKEN is empty, can't setup github client")
	}
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Println(LogEntry{
			Message:   fmt.Sprintf("failed create firestore client, error: %s", err.Error()),
			Severity:  "CRITICAL",
			Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
		})
		panic(fmt.Sprintf("Failed to create client, error: %s", err.Error()))
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv(githubAccessToken)},
	)
	tc := oauth2.NewClient(ctx, ts)

	githubClient = github.NewClient(tc)
}



for index, failureInstance := range failureInstances {
githubIssueNumber, err := failureInstance.DataAt("githubIssueNumber")
if err != nil {
log.Println(LogEntry{
Message:   "github issue for failing test doesn't exists",
Severity:  "INFO",
Trace:     trace,
Component: "kyma.prow.cloud-function.Getfailureinstancedetails",
Labels:    map[string]string{"messageId": contextMetadata.EventID, "jobID": jobID, "prowjobName": prowMessage.JobName},
})
} else {
check is issue closed
closed then close instance remove from array
}
}


for _, failureInstance := range failureInstances {
githubIssueNumber, err := failureInstance.DataAt("githubIssueNumber")
if err != nil {
log.Println("gh issue not found")
} else {
issue, _, err := githubClient.Issues.Get(ctx, githubOrg, githubRepo, githubIssueNumber.(int))
if err != nil {
log.Println(err.Error())
} else {
println(issue.State)
}
}
}
