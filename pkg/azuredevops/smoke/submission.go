package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
)

type buildTest struct {
	description  string
	logMessage   string
	expectAbsent bool
}

type timelineTest struct {
	name   string
	state  string
	result string
}

func main() {
	organizationUrl := os.Getenv("ORGANIZATION_URL")
	personalAccessToken := os.Getenv("PERSONAL_ACCESS_TOKEN")
	projectName := os.Getenv("PROJECT_NAME")
	pipelineName := os.Getenv("PIPELINE_NAME")
	pipelineId, err := strconv.Atoi(os.Getenv("TRIGGERED_PIPELINE_ID"))
	buildId, err := strconv.Atoi(os.Getenv("TRIGGERED_BUILD_ID"))

	if err != nil {
		log.Fatalf("Could not convert TRIGGERED_PIPELINE_ID to int: %s", err)
	}
	ctx := context.Background()

	connection := createPatConnection(organizationUrl, personalAccessToken)

	buildTests := []buildTest{
		{
			description:  "Checkout self repository",
			logMessage:   "Repository 'self' has been successfully checked out.",
			expectAbsent: false,
		},
		{
			description:  "Checkout kyma-modules repository",
			logMessage:   "Repository 'kyma-modules' has been successfully checked out.",
			expectAbsent: false,
		},
		{
			description:  "Checkout security-scans-modular repository",
			logMessage:   "Repository 'security-scans-modular' has been successfully checked out.",
			expectAbsent: false,
		},
		{
			description:  "Verify the download of conduit-cli",
			logMessage:   "conduit-cli has been successfully downloaded and is executable.",
			expectAbsent: false,
		},
		{
			description:  "Verify Python Installation",
			logMessage:   "command not found: python3",
			expectAbsent: true,
		},
		{
			description:  "Verify gcloud Installation",
			logMessage:   "command not found: gcloud",
			expectAbsent: true,
		},
		{
			description:  "Verify Go Installation",
			logMessage:   "command not found: go",
			expectAbsent: true,
		},
		{
			description:  "Download Kyma cli",
			logMessage:   "The 'kyma' binary has been successfully downloaded and is executable.",
			expectAbsent: false,
		},
	}

	for _, test := range buildTests {
		runBuildTest(ctx, connection, projectName, pipelineName, pipelineId, test)
	}
	timelineTests := []timelineTest{
		{
			name:   "Initialize job",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Install gcloud",
			state:  "completed",
			result: "succeeded",
		},
	}

	for _, test := range timelineTests {
		runTimelineTests(ctx, connection, projectName, buildId, test)
	}

}

func createPatConnection(organizationUrl, personalAccessToken string) *azuredevops.Connection {
	return azuredevops.NewPatConnection(organizationUrl, personalAccessToken)
}

func getSpecificBuilds(ctx context.Context, connection *azuredevops.Connection, projectName string, pipelineID int) ([]build.Build, error) {
	buildClient, err := build.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("error creating build client: %w", err)
	}

	// Args to download Build
	buildArgs := build.GetBuildsArgs{
		Project:     &projectName,
		Definitions: &[]int{pipelineID},
	}

	// Download Builds
	buildsResponse, err := buildClient.GetBuilds(ctx, buildArgs)
	if err != nil {
		return nil, fmt.Errorf("error getting builds: %w", err)
	}

	return buildsResponse.Value, nil
}

func getBuildStageStatus(ctx context.Context, connection *azuredevops.Connection, projectName string, buildId int, test timelineTest) (bool, error) {
	buildClient, err := build.NewClient(ctx, connection)
	if err != nil {
		return false, fmt.Errorf("error creating build client: %w", err)
	}

	// Args to download Build
	buildArgs := build.GetBuildTimelineArgs{
		Project: &projectName,
		BuildId: &buildId,
	}
	// Download Builds
	buildTimeline, err := buildClient.GetBuildTimeline(ctx, buildArgs)
	if err != nil {
		return false, fmt.Errorf("error getting builds: %w", err)
	}

	// Check each record
	for _, record := range *buildTimeline.Records {
		if record.Name != nil && *record.Name == test.name {
			if record.Result != nil && string(*record.Result) == test.result {
				if record.State != nil && string(*record.State) == test.state {
					return true, nil // Found a record matching all criteria
				} else {
					continue // The state doesn't match, continue checking other records
				}
			} else {
				continue // The result doesn't match, continue checking other records
			}
		}
		// Name doesn't match, continue checking other records
	}

	return false, fmt.Errorf("no record found matching the criteria")
}

func checkSpecificBuildForCommand(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName, logFinding string, pipelineId int) (bool, error) {
	builds, err := getSpecificBuilds(ctx, connection, projectName, pipelineId)
	if err != nil {
		return false, fmt.Errorf("error getting last build: %w", err)
	}
	if len(builds) == 0 {
		return false, fmt.Errorf("no builds found for pipeline %s", pipelineName)
	}

	buildClient, err := build.NewClient(ctx, connection)
	if err != nil {
		return false, fmt.Errorf("error creating build client: %w", err)
	}

	lastBuildID := builds[0].Id
	logs, err := buildClient.GetBuildLogs(ctx, build.GetBuildLogsArgs{
		Project: &projectName,
		BuildId: lastBuildID,
	})
	if err != nil {
		return false, fmt.Errorf("error getting build logs: %w", err)
	}

	// Search logs for usage of command `***`
	for _, log := range *logs {
		logContent, err := buildClient.GetBuildLogLines(ctx, build.GetBuildLogLinesArgs{
			Project: &projectName,
			BuildId: lastBuildID,
			LogId:   log.Id,
		})
		if err != nil {
			return false, fmt.Errorf("error getting build log lines: %w", err)
		}

		for _, line := range *logContent {
			if strings.Contains(line, logFinding) {
				return true, nil
			}
		}
	}

	return false, nil
}

func checkSpecificBuildForMissingCommand(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName, expectedMissingMessage string, pipelineId int) (bool, error) {
	builds, err := getSpecificBuilds(ctx, connection, projectName, pipelineId)
	if err != nil {
		return false, fmt.Errorf("error getting last build: %w", err)
	}
	if len(builds) == 0 {
		return false, fmt.Errorf("no builds found for pipeline %s", pipelineName)
	}

	buildClient, err := build.NewClient(ctx, connection)
	if err != nil {
		return false, fmt.Errorf("error creating build client: %w", err)
	}

	lastBuildID := builds[0].Id
	logs, err := buildClient.GetBuildLogs(ctx, build.GetBuildLogsArgs{
		Project: &projectName,
		BuildId: lastBuildID,
	})
	if err != nil {
		return false, fmt.Errorf("error getting build logs: %w", err)
	}

	// Search logs for usage of command `***`
	for _, log := range *logs {
		logContent, err := buildClient.GetBuildLogLines(ctx, build.GetBuildLogLinesArgs{
			Project: &projectName,
			BuildId: lastBuildID,
			LogId:   log.Id,
		})
		if err != nil {
			return false, fmt.Errorf("error getting build log lines: %w", err)
		}

		for _, line := range *logContent {
			if strings.Contains(line, expectedMissingMessage) {
				return false, fmt.Errorf("unexpected message found in logs: %s", expectedMissingMessage)
			}
		}
	}

	return true, nil
}

func runBuildTest(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName string, pipelineId int, test buildTest) bool {
	var pass bool
	var err error

	if test.expectAbsent {
		pass, err = checkSpecificBuildForMissingCommand(ctx, connection, projectName, pipelineName, test.logMessage, pipelineId)
	} else {
		pass, err = checkSpecificBuildForCommand(ctx, connection, projectName, pipelineName, test.logMessage, pipelineId)
	}

	if err != nil {
		fmt.Printf("Test failed for %s: %v\n", test.description, err)
		return false
	}

	if !pass {
		fmt.Printf("Test failed for %s: condition not met\n", test.description)
		return false
	}

	fmt.Printf("Test passed for %s\n", test.description)
	return true
}

func runTimelineTests(ctx context.Context, connection *azuredevops.Connection, projectName string, buildId int, test timelineTest) bool {
	var pass bool
	var err error

	pass, err = getBuildStageStatus(ctx, connection, projectName, buildId, test)
	if err != nil {
		fmt.Printf("Test failed for %s: %v\n", test.name, err)
		return false
	}

	if !pass {
		fmt.Printf("Test failed for %s: condition not met\n", test.name)
		return false
	}

	fmt.Printf("Test passed for %s\n", test.name)
	return true
}
