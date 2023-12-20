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
	"golang.org/x/exp/slices"
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
	// Fetching environment variables for Azure DevOps settings
	organizationUrl := os.Getenv("ORGANIZATION_URL")
	personalAccessToken := os.Getenv("PERSONAL_ACCESS_TOKEN")
	projectName := os.Getenv("PROJECT_NAME")
	pipelineName := os.Getenv("PIPELINE_NAME")
	pipelineIdStr := os.Getenv("PIPELINE_ID")
	buildIdStr := os.Getenv("BUILD_ID")

	// Converting variables from string to integer
	pipelineId, err := strconv.Atoi(pipelineIdStr)
	if err != nil {
		log.Fatalf("Error parsing PIPELINE_ID: %v", err)
	}
	buildId, err := strconv.Atoi(buildIdStr)
	if err != nil {
		log.Fatalf("Error parsing BUILD_ID: %v", err)
	}

	// Setting up context for API calls
	ctx := context.Background()

	// Creating a connection to Azure DevOps using the Personal Access Token
	connection := createPatConnection(organizationUrl, personalAccessToken)

	// Determining which tests to run based on the TESTS_TO_RUN environment variable
	testsToRun := os.Getenv("TESTS_TO_RUN")
	fmt.Println(testsToRun)
	var testsToRunList []string
	if testsToRun != "" && testsToRun != "all" {
		testsToRunList = strings.Split(testsToRun, ",")
	}

	// Defining build tests with their descriptions and expected log messages
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

	// Running each build test if it meets the criteria specified in TESTS_TO_RUN
	for _, test := range buildTests {
		if shouldRunTest(testsToRun, testsToRunList, test.description) {
			runBuildTest(ctx, connection, projectName, pipelineName, pipelineId, test)
		}
	}

	// Defining timeline tests with their names and expected states and results
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
		{
			name:   "Checkout kyma/module-manifests@main to s/module-manifests",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Checkout kyma/kyma-modules@main to s/kyma-modules",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Checkout kyma/security-scans-modular@main to s/security-scans-modular",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Download conduit-cli",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Install Python",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Install gcloud",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Install Go",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Collect module info",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Get SA token from Vault",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Save SA token to file",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Clone module repo",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Get Docker Password from GCP",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Download Kyma cli",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Create Module",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Dump module template as artifact",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Clean Up Pre-Submit Related Version",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Dump modulemanifest as artifact",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Push Moduletemplate to Main (kyma-modules)",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Clean Up Untagged Versions",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Post-job: Checkout kyma/kyma-modules@main to s/kyma-modules",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Post-job: Checkout kyma/module-manifests@main to s/module-manifests",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Finalize Job",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Post-job: Checkout kyma/security-scans-modular@main to s/security-scans-modular",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Post-Submit process - Publishing",
			state:  "completed",
			result: "succeeded",
		},
		{
			name:   "Post-Submit process - Publishing",
			state:  "completed",
			result: "succeeded",
		},
	}

	// Running each timeline test if it meets the criteria specified in TESTS_TO_RUN
	for _, test := range timelineTests {
		if shouldRunTest(testsToRun, testsToRunList, test.name) {
			runTimelineTests(ctx, connection, projectName, buildId, test)
		}
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

	return checkBuildRecords(buildTimeline, test.name, test.result, test.state)
}

func checkBuildRecords(timeline *build.Timeline, testName, testResult, testState string) (bool, error) {
	for _, record := range *timeline.Records {
		if record.Name != nil && *record.Name == testName {
			if record.Result != nil && string(*record.Result) == testResult && record.State != nil && string(*record.State) == testState {
				return true, nil // Found a record matching all criteria
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
		log.Fatalf("Test failed for %s: %v\n", test.description, err)
		return false
	}

	if !pass {
		log.Fatalf("Test failed for %s: condition not met\n", test.description)
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

var unittests = []struct {
	name           string
	timeline       *build.Timeline
	testName       string
	testResult     build.TaskResult
	testState      string
	expectedResult bool
	expectedError  bool
}{
	{
		name: "Record matches criteria",
		timeline: &build.Timeline{
			Records: &[]build.TimelineRecord{
				{
					Name:   strPtr("testName"),
					Result: taskResultPtr(build.TaskResultValues.Succeeded),
					State:  timelineRecordStatePtr(build.TimelineRecordStateValues.Completed),
				},
			},
		},
		testName:       "testName",
		testResult:     build.TaskResultValues.Succeeded,
		testState:      "completed",
		expectedResult: true,
	},
	{
		name: "No matching record",
		timeline: &build.Timeline{
			Records: &[]build.TimelineRecord{
				{
					Name:   strPtr("otherName"),
					Result: taskResultPtr(build.TaskResultValues.Failed),
					State:  timelineRecordStatePtr(build.TimelineRecordStateValues.InProgress),
				},
			},
		},
		testName:       "testName",
		testResult:     build.TaskResultValues.Succeeded,
		testState:      "completed",
		expectedResult: false,
		expectedError:  true,
	},
}

// Helper function to create a pointer to a string (to simplify test case setup)
func strPtr(s string) *string {
	return &s
}

func taskResultPtr(tr build.TaskResult) *build.TaskResult {
	return &tr
}

func timelineRecordStatePtr(trs build.TimelineRecordState) *build.TimelineRecordState {
	return &trs
}

func shouldRunTest(testsToRun string, testsToRunList []string, testName string) bool {
	if testsToRun == "all" || testsToRun == "" {
		return true
	}
	return slices.Contains(testsToRunList, testName)
}
