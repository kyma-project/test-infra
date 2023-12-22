package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	smoketests "github.com/kyma-project/test-infra/pkg/azuredevops/smoke"
)

func main() {
	// Fetching environment variables for Azure DevOps settings
	organizationURL := os.Getenv("ORGANIZATION_URL")
	personalAccessToken := os.Getenv("PERSONAL_ACCESS_TOKEN")
	projectName := os.Getenv("PROJECT_NAME")
	pipelineName := os.Getenv("PIPELINE_NAME")
	pipelineIDStr := os.Getenv("PIPELINE_ID")
	buildIDStr := os.Getenv("BUILD_ID")

	// Converting variables from string to integer
	pipelineID, err := strconv.Atoi(pipelineIDStr)
	if err != nil {
		log.Fatalf("Error parsing PIPELINE_ID: %v", err)
	}
	buildID, err := strconv.Atoi(buildIDStr)
	if err != nil {
		log.Fatalf("Error parsing BUILD_ID: %v", err)
	}

	// Setting up context for API calls
	ctx := context.Background()

	// Creating a connection to Azure DevOps using the Personal Access Token
	connection := smoketests.CreatePatConnection(organizationURL, personalAccessToken)

	// Determining which tests to run based on the TESTS_TO_RUN environment variable
	testsToRun := os.Getenv("TESTS_TO_RUN")

	var testsToRunList []string
	if testsToRun != "" && testsToRun != "all" {
		for _, test := range strings.Split(testsToRun, ",") {
			testsToRunList = append(testsToRunList, strings.TrimSpace(test))
		}
	}

	buildTests := smoketests.GetBuildTests()
	// Running each build test if it meets the criteria specified in TESTS_TO_RUN
	for _, test := range buildTests {
		if smoketests.ShouldRunTest(testsToRun, testsToRunList, test.Description) {
			smoketests.RunBuildTest(ctx, connection, projectName, pipelineName, pipelineID, test)
		}
	}

	timelineTests := smoketests.GetTimelineTests()
	// Running each timeline test if it meets the criteria specified in TESTS_TO_RUN
	for _, test := range timelineTests {
		if smoketests.ShouldRunTest(testsToRun, testsToRunList, test.Name) {
			smoketests.RunTimelineTests(ctx, connection, projectName, buildID, test)
		}
	}

}
