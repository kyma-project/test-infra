package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	gcplogging "github.com/kyma-project/test-infra/development/gcp/pkg/logging"
	"github.com/kyma-project/test-infra/development/github/pkg/client"
	"github.com/kyma-project/test-infra/development/prow"
	log "github.com/sirupsen/logrus"
)

func main() {
	var exitCode interface{}
	defer func() { os.Exit(exitCode.(int)) }()
	ctx := context.Background()
	var wg sync.WaitGroup
	saProwjobGcpLoggingClientKeyPath := os.Getenv("SA_PROWJOB_GCP_LOGGING_CLIENT_KEY_PATH")
	logClient, err := gcplogging.NewProwjobClient(ctx, saProwjobGcpLoggingClientKeyPath)
	if err != nil {
		log.Errorf("creating gcp logging client failed, got error: %v", err)
	}
	logger := logClient.NewProwjobLogger().WithGeneratedTrace()
	defer logger.Flush()
	// provided by preset-bot-github-sap-token
	accessToken := os.Getenv("BOT_GITHUB_SAP_TOKEN")
	contextLogger := logger.WithContext("checking if user exists in users map")
	defer contextLogger.Flush()
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
	wg.Add(len(authors))
	for _, author := range authors {
		go func(wg *sync.WaitGroup, author string, exitCode *int32) {
			defer wg.Done()
			for _, user := range usersMap {
				if user.ComGithubUsername == author {
					return
				}
			}
			contextLogger.LogError(fmt.Sprintf("user %s is not present in users map, please add user", author))
			atomic.StoreInt32(exitCode, 1)
		}(&wg, author, exitCode.(*int32))
	}
	wg.Wait()
	if exitCode == nil {
		contextLogger.LogInfo("all authors present in users map")
		exitCode = 0
	}
}
