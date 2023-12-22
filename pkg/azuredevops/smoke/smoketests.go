package smoketests

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"golang.org/x/exp/slices"
)

type BuildTest struct {
	Description  string
	LogMessage   string
	ExpectAbsent bool
}

type TimelineTest struct {
	Name   string
	State  string
	Result string
}

func CreatePatConnection(organizationURL, personalAccessToken string) *azuredevops.Connection {
	return azuredevops.NewPatConnection(organizationURL, personalAccessToken)
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

func getBuildStageStatus(ctx context.Context, connection *azuredevops.Connection, projectName string, buildID int, test TimelineTest) (bool, error) {
	buildClient, err := build.NewClient(ctx, connection)
	if err != nil {
		return false, fmt.Errorf("error creating build client: %w", err)
	}

	// Args to download Build
	buildArgs := build.GetBuildTimelineArgs{
		Project: &projectName,
		BuildId: &buildID,
	}
	// Download Builds
	buildTimeline, err := buildClient.GetBuildTimeline(ctx, buildArgs)
	if err != nil {
		return false, fmt.Errorf("error getting builds: %w", err)
	}

	return checkBuildRecords(buildTimeline, test.Name, test.Result, test.State)
}

func checkBuildRecords(timeline *build.Timeline, testName, testResult, testState string) (bool, error) {
	for _, record := range *timeline.Records {
		if record.Name != nil && *record.Name == testName {
			if record.Result != nil && string(*record.Result) == testResult && record.State != nil && string(*record.State) == testState {
				return true, nil // Found a record matching all criteria
			}
		}
	}

	return false, fmt.Errorf("no record found matching the criteria")
}

func checkSpecificBuildForCommand(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName, logFinding string, pipelineID int) (bool, error) {
	builds, err := getSpecificBuilds(ctx, connection, projectName, pipelineID)
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
	for _, buildLog := range *logs {
		logContent, err := buildClient.GetBuildLogLines(ctx, build.GetBuildLogLinesArgs{
			Project: &projectName,
			BuildId: lastBuildID,
			LogId:   buildLog.Id,
		})
		if err != nil {
			return false, fmt.Errorf("error getting build buildLog lines: %w", err)
		}

		for _, line := range *logContent {
			if strings.Contains(line, logFinding) {
				return true, nil
			}
		}
	}

	return false, nil
}

func checkSpecificBuildForMissingCommand(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName, expectedMissingMessage string, pipelineID int) (bool, error) {
	builds, err := getSpecificBuilds(ctx, connection, projectName, pipelineID)
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
	for _, buildLog := range *logs {
		logContent, err := buildClient.GetBuildLogLines(ctx, build.GetBuildLogLinesArgs{
			Project: &projectName,
			BuildId: lastBuildID,
			LogId:   buildLog.Id,
		})
		if err != nil {
			return false, fmt.Errorf("error getting build buildLog lines: %w", err)
		}

		for _, line := range *logContent {
			if strings.Contains(line, expectedMissingMessage) {
				return false, fmt.Errorf("unexpected message found in logs: %s", expectedMissingMessage)
			}
		}
	}

	return true, nil
}

func RunBuildTest(ctx context.Context, connection *azuredevops.Connection, projectName, pipelineName string, pipelineID int, test BuildTest) bool {
	var pass bool
	var err error

	if test.ExpectAbsent {
		pass, err = checkSpecificBuildForMissingCommand(ctx, connection, projectName, pipelineName, test.LogMessage, pipelineID)
	} else {
		pass, err = checkSpecificBuildForCommand(ctx, connection, projectName, pipelineName, test.LogMessage, pipelineID)
	}

	if err != nil {
		log.Fatalf("Test failed for %s: %v\n", test.Description, err)
	}

	if !pass {
		log.Fatalf("Test failed for %s: condition not met\n", test.Description)
	}

	fmt.Printf("Test passed for %s\n", test.Description)
	return true
}

func RunTimelineTests(ctx context.Context, connection *azuredevops.Connection, projectName string, buildID int, test TimelineTest) bool {
	var pass bool
	var err error

	pass, err = getBuildStageStatus(ctx, connection, projectName, buildID, test)
	if err != nil {
		log.Fatalf("Test failed for %s: %v\n", test.Name, err)
	}

	if !pass {
		log.Fatalf("Test failed for %s: condition not met\n", test.Name)
	}

	fmt.Printf("Test passed for %s\n", test.Name)
	return true
}

func ShouldRunTest(testsToRun string, testsToRunList []string, testName string) bool {
	if testsToRun == "all" || testsToRun == "" {
		return true
	}
	return slices.Contains(testsToRunList, testName)
}

func GetBuildTests() []BuildTest {
	return []BuildTest{
		{
			Description:  "Checkout self repository",
			LogMessage:   "Repository 'self' has been successfully checked out.",
			ExpectAbsent: false,
		},
		{
			Description:  "Checkout kyma-modules repository",
			LogMessage:   "Repository 'kyma-modules' has been successfully checked out.",
			ExpectAbsent: false,
		},
		{
			Description:  "Checkout security-scans-modular repository",
			LogMessage:   "Repository 'security-scans-modular' has been successfully checked out.",
			ExpectAbsent: false,
		},
		{
			Description:  "Verify the download of conduit-cli",
			LogMessage:   "conduit-cli has been successfully downloaded and is executable.",
			ExpectAbsent: false,
		},
		{
			Description:  "Verify Python Installation",
			LogMessage:   "command not found: python3",
			ExpectAbsent: true,
		},
		{
			Description:  "Verify gcloud Installation",
			LogMessage:   "command not found: gcloud",
			ExpectAbsent: true,
		},
		{
			Description:  "Verify Go Installation",
			LogMessage:   "command not found: go",
			ExpectAbsent: true,
		},
		{
			Description:  "Download Kyma cli",
			LogMessage:   "The 'kyma' binary has been successfully downloaded and is executable.",
			ExpectAbsent: false,
		},
	}
}

func GetTimelineTests() []TimelineTest {
	return []TimelineTest{
		{
			Name:   "Initialize job",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Install gcloud",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Checkout kyma/module-manifests@main to s/module-manifests",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Checkout kyma/kyma-modules@main to s/kyma-modules",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Checkout kyma/security-scans-modular@main to s/security-scans-modular",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Download conduit-cli",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Install Python",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Install gcloud",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Install Go",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Collect module info",
			State:  "completed",
			Result: "skipped",
		},
		{
			Name:   "Get SA token from Vault",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Save SA token to file",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Clone module repo",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Get Docker Password from GCP",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Download Kyma cli",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Create Module",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Dump module template as artifact",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Clean Up Pre-Submit Related Version",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Dump modulemanifest as artifact",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Push Moduletemplate to Main (kyma-modules)",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Clean Up Untagged Versions",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Post-job: Checkout kyma/kyma-modules@main to s/kyma-modules",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Post-job: Checkout kyma/module-manifests@main to s/module-manifests",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Finalize Job",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Post-job: Checkout kyma/security-scans-modular@main to s/security-scans-modular",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Post-Submit process - Publishing",
			State:  "completed",
			Result: "succeeded",
		},
		{
			Name:   "Post-Submit process - Publishing",
			State:  "completed",
			Result: "succeeded",
		},
	}

}
