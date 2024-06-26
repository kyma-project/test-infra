package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
)

func main() {
	// Fetching environment variables for Azure DevOps settings
	organizationURL := os.Getenv("ORGANIZATION_URL")
	personalAccessToken := os.Getenv("ADO_PAT")
	projectName := os.Getenv("PROJECT_NAME")
	pipelineName := os.Getenv("PIPELINE_NAME")
	pipelineIDStr := os.Getenv("PIPELINE_ID")
	buildIDStr := os.Getenv("BUILD_ID")
	retryDelayStr := os.Getenv("RETRY_DELAY")
	retryAttemptsStr := os.Getenv("RETRY_ATTEMPTS")

	// Converting variables from string to integer
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		log.Fatalf("Error parsing PIPELINE_ID: %v", err)
	}
	buildID, err := strconv.Atoi(buildIDStr)
	if err != nil {
		log.Fatalf("Error parsing BUILD_ID: %v", err)
	}
	retryAttempts := 3
	if retryAttemptsStr != "" {
		var err error
		retryAttempts, err = strconv.Atoi(retryAttemptsStr)
		if err != nil {
			log.Fatalf("Error parsing RETRY_ATTEMPTS: %v", err)
		}
	}
	retryDelay := 30 * time.Second
	if retryDelayStr != "" {
		var err error
		retryDelay, err = time.ParseDuration(retryDelayStr)
		if err != nil {
			log.Fatalf("Error parsing RETRY_DELAY: %v", err)
		}
	}
	retryStrategy := pipelines.RetryStrategy{
		Attempts: uint(retryAttempts),
		Delay:    retryDelay,
	}

	// Setting up context for API calls
	ctx := context.Background()

	// Creating a build client using the Personal Access Token
	buildClient, err := pipelines.NewBuildClient(organizationURL, personalAccessToken)
	if err != nil {
		log.Fatalf("Error creating build client: %v", err)
	}

	// Determining which tests to run based on the test-suite.yaml file
	testsToRun := os.Getenv("TESTS_TO_RUN_FILE_PATH")

	buildTests, timelineTests, err := pipelines.GetTestsDefinition(testsToRun)
	if err != nil {
		log.Printf("Failed to get test definitions: %v", err)
	}

	// Running each build test if it exists in YAML file
	for _, test := range buildTests {
		err := pipelines.RunBuildTests(ctx, buildClient, retryStrategy, projectName, pipelineName, pipelineID, &buildID, test)
		if err != nil {
			log.Printf("Error running build test: %v\n", err)
		}
	}

	// Running each timeline test if it exists in YAML file
	for _, test := range timelineTests {
		err := pipelines.RunTimelineTests(ctx, buildClient, retryStrategy, projectName, &buildID, test)
		if err != nil {
			log.Printf("Error running timeline test: %v\n", err)
		}
	}

}
