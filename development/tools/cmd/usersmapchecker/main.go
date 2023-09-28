package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/test-infra/development/github/pkg/client"
	"github.com/kyma-project/test-infra/development/prow"
	"github.com/kyma-project/test-infra/development/types"
)

// Example fields in gcp logging.
// logName: "projects/sap-kyma-prow/logs/stdout"
//
//	resource: {
//	  labels: {
//	    cluster_name: "trusted-workload-kyma-prow"
//	    container_name: "test"
//	    location: "europe-west3"
//	    namespace_name: "default"
//	    pod_name: "cbb59657-fa91-11eb-baea-4e9acc7ce5e6"
//	    project_id: "sap-kyma-prow"
//	  }
//	  type: "k8s_container"
//
//	labels: {
//	  compute.googleapis.com/resource_name: "gke-trusted-workload-k-high-cpu-16-32-c8294afe-skrq"
//	  k8s-pod/created-by-prow: "true"
//	  k8s-pod/event-GUID: "cb549a8a-fa91-11eb-80a9-35f1ac609512"
//	  k8s-pod/preset-build-main: "true"
//	  k8s-pod/preset-cluster-use-ssd: "true"
//	  k8s-pod/preset-cluster-version: "true"
//	  k8s-pod/preset-debug-commando-oom: "true"
//	  k8s-pod/preset-dind-enabled: "true"
//	  k8s-pod/preset-docker-push-repository-gke-integration: "true"
//	  k8s-pod/preset-gc-compute-envs: "true"
//	  k8s-pod/preset-gc-project-env: "true"
//	  k8s-pod/preset-gke-upgrade-post-job: "true"
//	  k8s-pod/preset-kyma-artifacts-bucket: "true"
//	  k8s-pod/preset-kyma-guard-bot-github-token: "true"
//	  k8s-pod/preset-log-collector-slack-token: "true"
//	  k8s-pod/preset-sa-gke-kyma-integration: "true"
//	  k8s-pod/preset-sa-test-gcr-push: "true"
//	  k8s-pod/prow_k8s_io/build-id: "1425409012446269440"
//	  k8s-pod/prow_k8s_io/context: "post-main-kyma-gke-upgrade"
//	  k8s-pod/prow_k8s_io/id: "cbb59657-fa91-11eb-baea-4e9acc7ce5e6"
//	  k8s-pod/prow_k8s_io/job: "post-main-kyma-gke-upgrade"
//	  k8s-pod/prow_k8s_io/plank-version: "v20210714-62f15287bd"
//	  k8s-pod/prow_k8s_io/pubsub_project: "sap-kyma-prow"
//	  k8s-pod/prow_k8s_io/pubsub_runID: "post-main-kyma-gke-upgrade"
//	  k8s-pod/prow_k8s_io/pubsub_topic: "prowjobs"
//	  k8s-pod/prow_k8s_io/refs_base_ref: "main"
//	  k8s-pod/prow_k8s_io/refs_org: "kyma-project"
//	  k8s-pod/prow_k8s_io/refs_repo: "kyma"
//	  k8s-pod/prow_k8s_io/type: "postsubmit"
//	}

// checkUserInMap is a function that checks if the author exists in the usersMap.
// It returns true if found and false otherwise.
func checkUserInMap(author string, usersMap []types.User) bool {
	for _, user := range usersMap {
		if user.ComGithubUsername == author {
			return true
		}
	}
	return false
}

func main() {
	ctx := context.Background()
	var missingUsers []string

	log.SetFormatter(&log.JSONFormatter{})
	// GitHub access token, provided by preset-bot-github-sap-token
	accessToken := os.Getenv("BOT_GITHUB_SAP_TOKEN")
	githubComAccessToken := os.Getenv("BOT_GITHUB_TOKEN")
	saptoolsClient, err := client.NewSapToolsClient(ctx, accessToken)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed creating sap tools github client, got error: %v", err))
	}

	githubComClient, err := client.NewClient(ctx, githubComAccessToken)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed creating github.com client, got error: %v", err))
	}
	usersMap, err := saptoolsClient.GetUsersMap(ctx)
	if err != nil {
		log.Fatalf(fmt.Sprintf("error when getting users map: got error %v", err))
	}
	authors, err := prow.GetPrAuthorForPresubmit()
	if err != nil {
		if notPresubmit := prow.IsNotPresubmitError(err); *notPresubmit {
			log.Infof(err.Error())
		} else {
			log.Fatalf(fmt.Sprintf("error when getting pr author for presubmit: got error %v", err))
		}
	}

	org, err := prow.GetOrgForPresubmit()
	if err != nil {
		if notPresubmit := prow.IsNotPresubmitError(err); *notPresubmit {
			log.Infof(err.Error())
		} else {
			log.Fatalf(fmt.Sprintf("error when getting org for presubmit: got error %v", err))
		}
	}

	log.Infof(fmt.Sprintf("found %d authors in job spec env variable", len(authors)))

	for _, author := range authors {
		// Check if author is a member of the organization.
		member, _, err := githubComClient.Organizations.IsMember(ctx, org, author)
		if err != nil {
			log.Fatalf(fmt.Sprintf("failed check if user %s is an github organisation member", author))
		}
		// If the author is a member of the organization but not present in usersMap, add to missingUsers.
		if member && !checkUserInMap(author, usersMap) {
			missingUsers = append(missingUsers, author)
		}
	}

	// If there are missing users, log a fatal error with all missing users, otherwise log an info message.
	if len(missingUsers) > 0 {
		log.Fatalf("users not present in users map: %v, please add them to users-map.yaml file.", missingUsers)
	}
	log.Infof("all authors present in users map")
}
