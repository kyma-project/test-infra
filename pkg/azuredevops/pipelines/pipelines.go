// Package pipelines provides a clients for calling Azure DevOps pipelines API
// TODO: Add more structured logging with debug severity to track execution in case of troubleshooting
package pipelines

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	adov7 "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
	"k8s.io/utils/ptr"
)

type Client interface {
	RunPipeline(ctx context.Context, args pipelines.RunPipelineArgs) (*pipelines.Run, error)
	GetRun(ctx context.Context, args pipelines.GetRunArgs) (*pipelines.Run, error)
}

type BuildClient interface {
	GetBuildLogs(ctx context.Context, args build.GetBuildLogsArgs) (*[]build.BuildLog, error)
	GetBuilds(ctx context.Context, args build.GetBuildsArgs) (*build.GetBuildsResponseValue, error)
	GetBuildLogLines(ctx context.Context, args build.GetBuildLogLinesArgs) (*[]string, error)
	GetBuildTimeline(ctx context.Context, args build.GetBuildTimelineArgs) (*build.Timeline, error)
}

type Tests struct {
	BuildTests    []BuildTest    `yaml:"buildTests"`
	TimelineTests []TimelineTest `yaml:"timelineTests"`
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

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

type Config struct {
	// ADO organization URL to call for triggering ADO pipeline
	ADOOrganizationURL string `yaml:"ado-organization-url" json:"ado-organization-url"`
	// ADO project name to call for triggering ADO pipeline
	ADOProjectName string `yaml:"ado-project-name" json:"ado-project-name"`
	// ADO pipeline ID to call for triggering ADO pipeline
	ADOPipelineID int `yaml:"ado-pipeline-id" json:"ado-pipeline-id"`
	// ADO pipeline version to call for triggering ADO pipeline
	ADOPipelineVersion int `yaml:"ado-pipeline-version,omitempty" json:"ado-pipeline-version,omitempty"`
}

func (c Config) GetADOConfig() Config {
	return c
}

// TODO: write tests which use BeAssignableToTypeOf matcher https://onsi.github.io/gomega/#beassignabletotypeofexpected-interface
func NewClient(adoOrganizationURL, adoPAT string) Client {
	adoConnection := adov7.NewPatConnection(adoOrganizationURL, adoPAT)
	ctx := context.Background()
	return pipelines.NewClient(ctx, adoConnection)
}

// TODO: write tests which use BeAssignableToTypeOf matcher https://onsi.github.io/gomega/#beassignabletotypeofexpected-interface
func NewBuildClient(adoOrganizationURL, adoPAT string) (BuildClient, error) {
	buildConnection := adov7.NewPatConnection(adoOrganizationURL, adoPAT)
	ctx := context.Background()
	return build.NewClient(ctx, buildConnection)
}

// TODO: implement sleep parameter to be passed as a functional option
func GetRunResult(ctx context.Context, adoClient Client, adoConfig Config, pipelineRunID *int, sleep time.Duration) (*pipelines.RunResult, error) {
	for {
		time.Sleep(sleep)
		pipelineRun, err := adoClient.GetRun(ctx, pipelines.GetRunArgs{
			Project:    &adoConfig.ADOProjectName,
			PipelineId: &adoConfig.ADOPipelineID,
			RunId:      pipelineRunID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed getting ADO pipeline run, err: %w", err)
		}
		if *pipelineRun.State == pipelines.RunStateValues.Completed {
			return pipelineRun.Result, nil
		}
		// TODO: use structured logging with info severity
		fmt.Printf("Pipeline run still in progress. Waiting for %s\n", sleep)
	}
}

func GetRunLogs(ctx context.Context, buildClient BuildClient, httpClient HTTPClient, adoConfig Config, pipelineRunID *int, adoPAT string) (string, error) {
	buildLogs, err := buildClient.GetBuildLogs(ctx, build.GetBuildLogsArgs{
		Project: &adoConfig.ADOProjectName,
		BuildId: pipelineRunID,
	})
	if err != nil {
		return "", fmt.Errorf("failed getting build logs metadata, err: %w", err)
	}

	// Last item in a list represent logs from all pipeline steps visible in ADO GUI
	lastLog := (*buildLogs)[len(*buildLogs)-1]
	req, err := http.NewRequest("GET", *lastLog.Url, nil)
	if err != nil {
		return "", fmt.Errorf("failed creating http request getting build log, err: %w", err)
	}
	req.SetBasicAuth("", adoPAT)
	// TODO: implement checking http response status code, if it's not 2xx, return error
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed http request getting build log, err: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed reading http body with build log, err: %w", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("failed closing http body with build log, err: %w", err)
	}
	return string(body), nil
}

// GetBuildStageStatus retrieves the status of a specific stage in a build process, based on the criteria defined in a TimelineTest.
// It first fetches the build timeline for a given build ID and then checks for a record in the timeline that matches the test criteria.
//
// Parameters:
// ctx          - The context to control the execution and cancellation of the test.
// buildClient  - The client interface to interact with the build system.
// pipelineName - The name of the pipeline within the project.
// buildID      - A pointer to an integer storing the build identifier.
// test         - The TimelineTest struct containing the name, expected result, and state of the test stage to be checked.
//
// Returns a boolean and an error. The boolean is true if a record matching the test's criteria (name, result, and state)
// is found in the build timeline. If no matching record is found or if there is an error in retrieving the build timeline,
// the function returns false and an error with a detailed message for troubleshooting.
//
// This function is particularly useful for verifying specific stages or conditions in a build process, especially in continuous
// integration and deployment scenarios where automated verification of build stages is required.
func GetBuildStageStatus(ctx context.Context, buildClient BuildClient, projectName string, buildID *int, test TimelineTest) (bool, error) {
	buildArgs := build.GetBuildTimelineArgs{
		Project: &projectName,
		BuildId: buildID,
	}

	buildTimeline, err := buildClient.GetBuildTimeline(ctx, buildArgs)
	if err != nil {
		return false, fmt.Errorf("error getting build timeline: %w", err)
	}

	return CheckBuildRecords(buildTimeline, test.Name, test.Result, test.State)
}

// CheckSpecificBuildForCommand searches for the occurrence of a specified command or message within the build logs of a specific build.
// This function is used to verify if a particular command or log entry was executed or generated during the build process.
//
// Parameters:
// ctx          - The context to control the execution and cancellation of the test.
// buildClient  - The client interface to interact with the build system.
// buildID      - A pointer to an integer storing the build identifier.
// projectName  - The name of the project in which the test is being run.
// pipelineName - The name of the pipeline within the project.
// logFinding   - The specific command or message to search for in the build logs.
// pipelineID   - The identifier of the pipeline.
//
// Returns a boolean and an error. The boolean is true if the specified command or message is found in the build logs,
// indicating that the command was executed. It returns false if the message is not found or if there are no logs for the build.
// In case of an error in fetching builds or logs, or any other operational issue, the function returns the error with a detailed
// message for troubleshooting.
func CheckSpecificBuildForCommand(ctx context.Context, buildClient BuildClient, projectName, pipelineName, logFinding string, pipelineID int, buildID *int) (bool, error) {
	buildArgs := build.GetBuildsArgs{
		Project:     &projectName,
		Definitions: &[]int{pipelineID},
	}
	buildsResponse, err := buildClient.GetBuilds(ctx, buildArgs)
	if err != nil {
		return false, fmt.Errorf("error getting last build: %w", err)
	}

	if len(buildsResponse.Value) == 0 {
		return false, fmt.Errorf("no builds found for pipeline %s", pipelineName)
	}

	logs, err := buildClient.GetBuildLogs(ctx, build.GetBuildLogsArgs{
		Project: &projectName,
		BuildId: buildID,
	})
	if err != nil {
		return false, fmt.Errorf("error getting build logs: %w", err)
	}

	// Search logs for usage of command `logFinding`
	for _, buildLog := range *logs {
		logContent, err := buildClient.GetBuildLogLines(ctx, build.GetBuildLogLinesArgs{
			Project: &projectName,
			BuildId: buildID,
			LogId:   buildLog.Id,
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

// CheckSpecificBuildForMissingCommand verifies if a specified message is absent in the build logs of a given build.
// It is primarily used to ensure that a certain command or log message was not executed or generated during the build process.
//
// Parameters:
// ctx                    - The context to control the execution and cancellation of the test.
// buildClient            - The client interface to interact with the build system.
// buildID                - A pointer to an integer storing the build identifier.
// projectName            - The name of the project in which the test is being run.
// pipelineName           - The name of the pipeline within the project.
// expectedMissingMessage - The message or command that is expected to be absent in the build logs.
// pipelineID    		  - The identifier of the pipeline.
//
// Returns a boolean and an error. The boolean is true if the specified message is indeed missing from the build logs,
// indicating that the command was not executed. It returns false if the message is found or if there are no logs for the build.
// In case of an error in fetching builds or logs, or any other operational issue, the function returns the error with a detailed
// message for troubleshooting.
func CheckSpecificBuildForMissingCommand(ctx context.Context, buildClient BuildClient, buildID *int, projectName, pipelineName, expectedMissingMessage string, pipelineID int) (bool, error) {
	buildArgs := build.GetBuildsArgs{
		Project:     &projectName,
		Definitions: &[]int{pipelineID},
	}
	buildsResponse, err := buildClient.GetBuilds(ctx, buildArgs)
	if err != nil {
		return false, fmt.Errorf("error getting last build: %w", err)
	}

	if len(buildsResponse.Value) == 0 {
		return false, fmt.Errorf("no builds found for pipeline %s", pipelineName)
	}

	logs, err := buildClient.GetBuildLogs(ctx, build.GetBuildLogsArgs{
		Project: &projectName,
		BuildId: buildID,
	})
	if err != nil {
		return false, fmt.Errorf("error getting build logs: %w", err)
	}

	// Search logs for the expected missing message
	for _, buildLog := range *logs {
		logContent, err := buildClient.GetBuildLogLines(ctx, build.GetBuildLogLinesArgs{
			Project: &projectName,
			BuildId: buildID,
			LogId:   buildLog.Id,
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

// CheckBuildRecords examines a build timeline to find a specific test record that matches the given criteria.
// It searches for a record within the timeline that has the specified test name, result, and state.
//
// Parameters:
// timeline  - A pointer to a build.Timeline struct containing a slice of build records.
// testName  - The name of the test to look for within the build records.
// testResult - The expected result of the test (e.g., "Succeeded", "Skipped").
// testState  - The expected state of the test (e.g., "Completed", "Pending").
//
// Returns a boolean and an error. The boolean is true if a record matching all the specified criteria (test name, result, and state)
// is found in the timeline. If no matching record is found, the function returns false and an error indicating the absence of a
// record that meets the specified conditions.
//
// This function is useful for verifying specific outcomes in a series of build tests, particularly for continuous integration and
// deployment scenarios where test results need to be programmatically verified.
func CheckBuildRecords(timeline *build.Timeline, testName, testResult, testState string) (bool, error) {
	for _, record := range *timeline.Records {
		if record.Name != nil && *record.Name == testName {
			if record.Result != nil && string(*record.Result) == testResult && record.State != nil && string(*record.State) == testState {
				return true, nil // Found a record matching all criteria
			}
		}
	}

	return false, fmt.Errorf("no record found matching the criteria")
}

// RunBuildTest executes a build test within a given context. It uses the specified build client
// to run tests on a project and pipeline, based on the provided pipeline and build IDs.
// It evaluates the test condition (presence or absence of a specific command) in the build logs.
//
// Parameters:
// ctx           - The context to control the execution and cancellation of the test.
// buildClient   - The client interface to interact with the build system.
// projectName   - The name of the project in which the test is being run.
// pipelineName  - The name of the pipeline within the project.
// pipelineID    - The identifier of the pipeline.
// buildID       - A pointer to an integer storing the build identifier.
// test          - The build test to be executed, which includes test conditions and expectations.
//
// Returns true if the test passes, which includes successful execution and meeting of the test conditions.
// If the test fails due to an error in execution or if the test conditions are not met, the function
// logs a fatal error with the test description and the reason for the failure.
func RunBuildTest(ctx context.Context, buildClient BuildClient, projectName, pipelineName string, pipelineID int, buildID *int, test BuildTest) bool {
	var pass bool
	var err error

	if test.ExpectAbsent {
		pass, err = CheckSpecificBuildForMissingCommand(ctx, buildClient, buildID, projectName, pipelineName, test.LogMessage, pipelineID)
	} else {
		pass, err = CheckSpecificBuildForCommand(ctx, buildClient, projectName, pipelineName, test.LogMessage, pipelineID, buildID)
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

// RunTimelineTests conducts a series of tests based on the timeline of a build process.
// It checks the status of different stages in the build process against the expectations
// defined in the TimelineTest structure.
//
// Parameters:
// ctx           - The context to control the execution and cancellation of the test.
// buildClient   - The client interface to interact with the build system.
// projectName   - The name of the project in which the test is being run.
// buildID       - A pointer to an integer storing the build identifier.
// test          - The build test to be executed, which includes test conditions and expectations.
//
// Returns true if the test passes, which includes successful execution and meeting of the test conditions.
// If the test fails due to an error in execution or if the test conditions are not met, the function
// logs a fatal error with the test description and the reason for the failure.
func RunTimelineTests(ctx context.Context, buildClient BuildClient, projectName string, buildID *int, test TimelineTest) bool {
	var pass bool
	var err error

	pass, err = GetBuildStageStatus(ctx, buildClient, projectName, buildID, test)
	if err != nil {
		log.Fatalf("Test failed for %s: %v\n", test.Name, err)
	}

	if !pass {
		log.Fatalf("Test failed for %s: condition not met\n", test.Name)
	}

	fmt.Printf("Test passed for %s\n", test.Name)
	return true
}

// GetTestsDefinition reads a YAML file from a specified path and unmarshalls it into slices of BuildTest and TimelineTest.
// This function is used for parsing test definitions from a YAML configuration file, allowing for dynamic test specifications.
//
// Parameters:
// filePath - The path to the YAML file that contains the test definitions.
//
// Returns two slices: one of BuildTest and another of TimelineTest. Each slice contains the respective test definitions
// extracted from the YAML file. The BuildTest slice contains tests related to build processes, whereas the TimelineTest
// slice contains tests that pertain to timeline events in a build.
//
// In case of errors in reading the file or unmarshalling the content, the function logs a fatal error with the specific issue
// and terminates the execution. This design assumes that the successful parsing of test definitions is critical for the
// continuation of the program.
func GetTestsDefinition(filePath string) (buildTests []BuildTest, timelineTests []TimelineTest) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("error reading tests file: %v", err)
	}

	var tests Tests
	err = yaml.Unmarshal(fileContent, &tests)
	if err != nil {
		log.Fatalf("error unmarshalling tests: %v", err)
	}

	return tests.BuildTests, tests.TimelineTests
}

func NewRunPipelineArgs(templateParameters map[string]string, adoConfig Config, pipelineRunArgs ...RunPipelineArgsOptions) (pipelines.RunPipelineArgs, error) {
	adoRunPipelineArgs := pipelines.RunPipelineArgs{
		Project:    &adoConfig.ADOProjectName,
		PipelineId: &adoConfig.ADOPipelineID,
		RunParameters: &pipelines.RunPipelineParameters{
			PreviewRun:         ptr.To(false),
			TemplateParameters: &templateParameters,
		},
	}
	if adoConfig.ADOPipelineVersion != 0 {
		adoRunPipelineArgs.PipelineVersion = &adoConfig.ADOPipelineVersion
	}
	for _, arg := range pipelineRunArgs {
		err := arg(&adoRunPipelineArgs)
		if err != nil {
			return pipelines.RunPipelineArgs{}, fmt.Errorf("failed setting pipeline run args, err: %w", err)
		}
	}
	// TODO: use structured logging with debug severity
	// fmt.Printf("Using TemplateParameters: %+v\n", adoRunPipelineArgs.RunParameters.TemplateParameters)
	return adoRunPipelineArgs, nil
}

type RunPipelineArgsOptions func(*pipelines.RunPipelineArgs) error

func PipelinePreviewRun(overrideYamlPath string) func(args *pipelines.RunPipelineArgs) error {
	return func(args *pipelines.RunPipelineArgs) error {
		args.RunParameters.PreviewRun = ptr.To(true)
		data, err := os.ReadFile(overrideYamlPath)
		if err != nil {
			return fmt.Errorf("failed reading override yaml file, err: %w", err)
		}
		overrideYaml := string(data)
		args.RunParameters.YamlOverride = &overrideYaml
		return nil
	}
}
