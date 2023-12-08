package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/pipelines"
)

const (
	organizationUrl     = "https://dev.azure.com/hyperspace-pipelines"
	personalAccessToken = "eintmxeevtckokpmpxxpxjbpeisd53qukafsymotkqkrjpb4akqq"
	projectName         = "kyma"
	pipelineName        = "module-manifests"
)

type buildTest struct {
	description  string
	logMessage   string
	expectAbsent bool
}

func main() {
	ctx := context.Background()

	connection := createPatConnection(organizationUrl, personalAccessToken)

	pipelineId, err := getPipelineID(ctx, connection, projectName, pipelineName)
	if err != nil {
		log.Fatalf("Error getting pipeline ID: %v", err)
	}

	fmt.Println(pipelineId)

	tests := []buildTest{
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

	for _, test := range tests {
		runBuildTest(ctx, connection, projectName, pipelineName, test)
	}
}

func createPatConnection(organizationUrl, personalAccessToken string) *azuredevops.Connection {
	return azuredevops.NewPatConnection(organizationUrl, personalAccessToken)
}

func getPipelineID(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName string) (int, error) {
	pipelineClient := pipelines.NewClient(ctx, connection)

	pipelinesResponse, err := pipelineClient.ListPipelines(ctx, pipelines.ListPipelinesArgs{Project: &projectName})
	if err != nil {
		return 0, fmt.Errorf("error listing pipelines: %w", err)
	}

	for _, pipeline := range pipelinesResponse.Value {
		if *pipeline.Name == pipelineName {
			return *pipeline.Id, nil
		}
	}

	return 0, fmt.Errorf("pipeline not found: %s", pipelineName)
}

func getLastBuilds(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName string, top int) ([]build.Build, error) {
	buildClient, err := build.NewClient(ctx, connection)
	if err != nil {
		return nil, fmt.Errorf("error creating build client: %w", err)
	}

	// Get Pipeline ID from pipeline name
	pipelineID, err := getPipelineID(ctx, connection, projectName, pipelineName)
	if err != nil {
		return nil, fmt.Errorf("error getting pipeline ID: %w", err)
	}

	// Args to download Build
	buildArgs := build.GetBuildsArgs{
		Project:     &projectName,
		Definitions: &[]int{pipelineID},
		Top:         &top,
	}

	// Download Builds
	buildsResponse, err := buildClient.GetBuilds(ctx, buildArgs)
	if err != nil {
		return nil, fmt.Errorf("error getting builds: %w", err)
	}

	return buildsResponse.Value, nil
}

func checkLastBuildForCommand(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName, logFinding string) (bool, error) {
	builds, err := getLastBuilds(ctx, connection, projectName, pipelineName, 1)
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

func checkLastBuildForMissingCommand(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName, expectedMissingMessage string) (bool, error) {
	builds, err := getLastBuilds(ctx, connection, projectName, pipelineName, 1)
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

func runBuildTest(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName string, test buildTest) bool {
	var pass bool
	var err error

	if test.expectAbsent {
		pass, err = checkLastBuildForMissingCommand(ctx, connection, projectName, pipelineName, test.logMessage)
	} else {
		pass, err = checkLastBuildForCommand(ctx, connection, projectName, pipelineName, test.logMessage)
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
