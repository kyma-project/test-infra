package main

import (
	"context"
	"fmt"
	gcplogging "github.com/kyma-project/test-infra/development/gcp/pkg/logging"
	"github.com/kyma-project/test-infra/development/github/pkg/client"
	"github.com/kyma-project/test-infra/development/prow"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

func main() {
	ctx := context.Background()
	var wg sync.WaitGroup
	saProwjobGcpLoggingClientKeyPath := os.Getenv("SA_PROWJOB_GCP_LOGGING_CLIENT_KEY_PATH")
	logClient, err := gcplogging.NewProwjobClient(ctx, saProwjobGcpLoggingClientKeyPath)
	if err != nil {
		log.Errorf("creating gcp logging client failed, got error: %v", err)
	}
	logger := logClient.NewProwjobLogger().WithGeneratedTrace()
	// provided by preset-bot-github-sap-token
	accessToken := os.Getenv("BOT_GITHUB_SAP_TOKEN")
	contextLogger := logger.WithContext("checking if user exists in users map")
	saptoolsClient, err := client.NewSapToolsClient(ctx, accessToken)
	if err != nil {
		contextLogger.LogError(fmt.Sprintf("failed creating sap tools github client, got error: %v", err))
	}
	usersMap, err := saptoolsClient.GetUsersMap(ctx)
	if err != nil {
		contextLogger.LogError(fmt.Sprintf("error when getting users map: got error %v", err))
	}
	authors, err := prow.GetPrAuthorForPresubmit()
	if err != nil {
		if notPresubmit := prow.IsNotPresubmitError(err); *notPresubmit {
			contextLogger.LogInfo(err.Error())
		} else {
			contextLogger.LogError(fmt.Sprintf("error when getting pr author for presubmit: got error %v", err))
		}
	}
	for _, author := range authors {
		go func(wg *sync.WaitGroup, author string) {
			defer wg.Done()
			wg.Add(1)
			for _, user := range usersMap {
				if user.ComGithubUsername == author {
					return
				}
			}
			contextLogger.LogError(fmt.Sprintf("user %s is not present in users map, please add user", author))
			os.Exit(1)
		}(&wg, author)
	}
	wg.Wait()
	contextLogger.LogInfo("all authors present in users map")
}
