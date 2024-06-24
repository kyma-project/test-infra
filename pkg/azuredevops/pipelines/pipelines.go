// Package pipelines allows calling Azure DevOps pipelines API to interact with kyma-project pipelines.
// It provides a set of functions to trigger a pipeline, get its status, and check the logs.
// It also includes functions to run tests on the build logs and timeline.
// These functions are designed to interact with kyma-project pipelines and it's tests.
// TODO: Add more structured logging with debug severity to track execution in case of troubleshooting
package pipelines

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	adov7 "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
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

// Retry strategy contains configuration for ADO request retry policy
//
// Fields:
// Attempts - Attempts to try before fail request
// Delay - Waiting time between next try
type RetryStrategy struct {
	// Attempts to make before failing request
	Attempts uint `yaml:"attempts" json:"attempts"`
	// Delay between two request tries
	Delay time.Duration `yaml:"delay" json:"delay"`
}

// Config is a struct that holds the configuration for Azure DevOps (ADO) pipelines.
// It includes the ADO organization URL, project name, pipeline ID, test pipeline ID, and pipeline version.
// These fields are used to trigger ADO pipelines and are required for the correct operation of the pipelines.
//
// Fields:
// ADOOrganizationURL: The URL of the ADO organization.
// ADOProjectName: The name of the ADO project.
// ADOPipelineID: The ID of the ADO oci-image-builder pipeline.
// ADOTestPipelineID: The ID of the ADO test pipeline.
// ADOPipelineVersion: The version of the ADO pipeline.
// ADORequestStrategy: Strategy for retrying failed requests to ADO API
type Config struct {
	// ADO organization URL to call for triggering ADO pipeline
	ADOOrganizationURL string `yaml:"ado-organization-url" json:"ado-organization-url"`
	// ADO project name to call for triggering ADO pipeline
	ADOProjectName string `yaml:"ado-project-name" json:"ado-project-name"`
	// ADO pipeline ID to call for triggering ADO pipeline
	ADOPipelineID int `yaml:"ado-pipeline-id" json:"ado-pipeline-id"`
	// ADO pipeline version to call for triggering ADO pipeline
	ADOPipelineVersion int `yaml:"ado-pipeline-version,omitempty" json:"ado-pipeline-version,omitempty"`
	// ADO Retry strategy for requests
	ADORetryStrategy RetryStrategy `yaml:"ado-retry-strategy" json:"ado-retry-strategy"`
}

func (c Config) GetADOConfig() Config {
	return c
}

// NewClient creates a new Azure DevOps (ADO) client to interact with ADO pipelines.
// It takes the ADO organization URL and a personal access token (PAT) as input parameters.
// Parameters:
// adoOrganizationURL - is the URL of the ADO organization containing the pipelines.
// adoPAT - is the personal access token for authentication in ADO API.
// TODO: write tests which use BeAssignableToTypeOf matcher https://onsi.github.io/gomega/#beassignabletotypeofexpected-interface
func NewClient(adoOrganizationURL, adoPAT string) Client {
	adoConnection := adov7.NewPatConnection(adoOrganizationURL, adoPAT)
	ctx := context.Background()
	return pipelines.NewClient(ctx, adoConnection)
}

// NewBuildClient creates a new Azure DevOps (ADO) build client to interact with ADO pipelines.
// Build client is used to get build logs and timeline for a specific pipeline run.
// It takes the ADO organization URL and a personal access token (PAT) as input parameters.
// Parameters:
// adoOrganizationURL - is the URL of the ADO organization containing the pipelines.
// adoPAT - is the personal access token for authentication in ADO API.
// TODO: write tests which use BeAssignableToTypeOf matcher https://onsi.github.io/gomega/#beassignabletotypeofexpected-interface
func NewBuildClient(adoOrganizationURL, adoPAT string) (BuildClient, error) {
	buildConnection := adov7.NewPatConnection(adoOrganizationURL, adoPAT)
	ctx := context.Background()
	return build.NewClient(ctx, buildConnection)
}

// GetRunResult is a function that retrieves the result of a specific Azure DevOps (ADO) pipeline run.
// It continuously checks the state of the pipeline run until it is completed.
// The function takes a context, an ADO client, an ADO configuration, a pipeline run ID, and a sleep duration as arguments.
//
// Parameters:
// adoClient - The ADO client to interact with the ADO pipelines.
// adoConfig - The configuration for the ADO pipeline organization.
// pipelineRunID - The ID of the pipeline run whose result is to be fetched.
// sleep - The duration to wait between each check of the pipeline run state.
//
// The function returns the result of the pipeline run and an error. If the pipeline run is still in progress,
// the function waits for the specified sleep duration before checking again. If an error occurs while getting
// the pipeline run, the function returns the error. If the pipeline run is completed, the function returns
// the result of the pipeline run.
// TODO: implement sleep parameter to be passed as a functional option
func GetRunResult(ctx context.Context, adoClient Client, adoConfig Config, pipelineRunID *int, sleep time.Duration) (*pipelines.RunResult, error) {
	for {
		// Sleep for the specified duration before checking the pipeline run state.
		time.Sleep(sleep)
		// Get the pipeline run. If an error occurs, retry three times with a delay of 5 seconds between each retry.
		// We get the pipeline run status over network, so we need to handle network errors.
		pipelineRun, err := retry.DoWithData[*pipelines.Run](
			func() (*pipelines.Run, error) {
				pipelineRun, err := adoClient.GetRun(ctx, pipelines.GetRunArgs{
					Project:    &adoConfig.ADOProjectName,
					PipelineId: &adoConfig.ADOPipelineID,
					RunId:      pipelineRunID,
				})
				return pipelineRun, err
			},
			retry.Attempts(adoConfig.ADORetryStrategy.Attempts),
			retry.Delay(adoConfig.ADORetryStrategy.Delay),
		)
		if err != nil {
			return nil, fmt.Errorf("failed getting ADO pipeline run, err: %w", err)
		}
		// If the pipeline run is completed, return the result of the pipeline run.
		if *pipelineRun.State == pipelines.RunStateValues.Completed {
			return pipelineRun.Result, nil
		}
		// If the pipeline run is still in progress, print a message and continue the loop.
		// TODO: use structured logging with info severity
		fmt.Printf("Pipeline run still in progress. Waiting for %s\n", sleep)
	}
}

// GetRunLogs is a function that retrieves the logs of a specific Azure DevOps (ADO) pipeline run.
// It fetches the build logs metadata and then makes an HTTP request to get the actual logs.
// The function takes a context, a build client, an HTTP client, an ADO configuration, a pipeline run ID, and an ADO personal access token (PAT) as arguments.
//
// Parameters:
// buildClient - The ADO build client to interact with the ADO pipelines.
// httpClient - The HTTP client to make requests to the ADO API to get logs content.
// adoConfig - The configuration for the ADO pipelines organization.
// pipelineRunID - The ID of the pipeline run whose logs are to be fetched.
// adoPAT - The personal access token for authentication in ADO API.
//
// The function returns the logs of the pipeline run as a string and an error. If an error occurs while getting
// the build logs metadata, making the HTTP request, reading the HTTP response body, or closing the HTTP response body,
// the function returns the error. If the pipeline run logs are successfully fetched, the function returns the logs.
func GetRunLogs(ctx context.Context, buildClient BuildClient, httpClient HTTPClient, adoConfig Config, pipelineRunID *int, adoPAT string) (string, error) {
	// Fetch the build logs metadata for the pipeline run. If an error occurs, retry three times with a delay of 5 seconds between each retry.
	// We get the pipeline run status over network, so we need to handle network errors.
	buildLogs, err := retry.DoWithData(
		func() (*[]build.BuildLog, error) {
			buildLogs, err := buildClient.GetBuildLogs(ctx, build.GetBuildLogsArgs{
				Project: &adoConfig.ADOProjectName,
				BuildId: pipelineRunID,
			})
			return buildLogs, err
		},
		retry.Attempts(adoConfig.ADORetryStrategy.Attempts),
		retry.Delay(adoConfig.ADORetryStrategy.Delay),
	)
	if err != nil {
		return "", fmt.Errorf("failed getting build logs metadata, err: %w", err)
	}

	// The last item in the list represents logs from all pipeline steps visible in ADO GUI
	lastLog := (*buildLogs)[len(*buildLogs)-1]
	// Create an HTTP request to get the actual logs content.
	req, err := http.NewRequest("GET", *lastLog.Url, nil)
	if err != nil {
		return "", fmt.Errorf("failed creating http request getting build log, err: %w", err)
	}
	req.SetBasicAuth("", adoPAT)
	// Make the HTTP request to get the actual logs content. If an error occurs, retry three times with a delay of 5 seconds between each retry.
	// We get the pipeline run status over network, so we need to handle network errors.
	// TODO: implement checking http response status code, if it's not 2xx, return error
	resp, err := retry.DoWithData[*http.Response](
		func() (*http.Response, error) {
			return httpClient.Do(req)
		},
		retry.Attempts(adoConfig.ADORetryStrategy.Attempts),
		retry.Delay(adoConfig.ADORetryStrategy.Delay),
	)
	if err != nil {
		return "", fmt.Errorf("failed http request getting build log, err: %w", err)
	}
	// Read the HTTP response body to get the logs content.
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

// CheckBuildLogForMessage verifies the presence or absence of a specified message in the build logs.
// It can be used to ensure whether a certain command or log message was executed/generated or not during the build process.
//
// Parameters:
// ctx           - The context to control the execution and cancellation of the test.
// buildClient   - The client interface to interact with the build system.
// projectName   - The name of the project in which the test is being run.
// pipelineName  - The name of the pipeline within the project.
// logMessage    - The specific command or message to search for in the build logs.
// expectAbsent  - Boolean indicating whether the message is expected to be absent (true) or present (false) in the logs.
// pipelineID    - The identifier of the pipeline.
// buildID       - A pointer to an integer storing the build identifier.
//
// Returns a boolean and an error. The boolean is true if the condition (presence or absence) of the specified message is met in the build logs.
// In case of an error in fetching builds or logs, or any other operational issue, the function returns the error with a detailed message for troubleshooting.
func CheckBuildLogForMessage(ctx context.Context, buildClient BuildClient, projectName, pipelineName, logMessage string, expectAbsent bool, pipelineID int, buildID *int) (bool, error) {
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
			found := strings.Contains(line, logMessage)
			if expectAbsent && found {
				return false, fmt.Errorf("unexpected message found in logs: %s", logMessage)
			} else if !expectAbsent && found {
				return true, nil
			}
		}
	}

	if expectAbsent {
		return true, nil // Message was correctly absent
	}
	return false, fmt.Errorf("message not found in logs: %s", logMessage)
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

// RunBuildTests executes a build test within a given context. It uses the specified build client
// to run tests on a project and pipeline, based on the provided pipeline and build IDs.
// This function checks the build logs to evaluate the test condition, which involves verifying
// the presence or absence of a specific command or message as defined in the test.
//
// Parameters:
// ctx           - The context to control the execution and cancellation of the test.
// buildClient   - The client interface to interact with the build system.
// projectName   - The name of the project in which the test is being run.
// pipelineName  - The name of the pipeline within the project.
// pipelineID    - The identifier of the pipeline.
// buildID       - A pointer to an integer storing the build identifier.
// test          - The build test to be executed, which includes test conditions (expecting the presence or absence of a log message) and expectations.
//
// Returns an error if the test fails due to an error in execution or if the test conditions (presence or absence of the specified log message) are not met.
// If the test passes, which includes successful execution and meeting of the test conditions, the function returns nil.
func RunBuildTests(ctx context.Context, buildClient BuildClient, projectName, pipelineName string, pipelineID int, buildID *int, test BuildTest) error {
	pass, err := CheckBuildLogForMessage(ctx, buildClient, projectName, pipelineName, test.LogMessage, test.ExpectAbsent, pipelineID, buildID)
	if err != nil {
		return fmt.Errorf("test failed for %s: %v", test.Description, err)
	}

	if !pass {
		return fmt.Errorf("test failed for %s: condition not met", test.Description)
	}

	fmt.Printf("Test passed for %s\n", test.Description)
	return nil
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
// test          - The timeline test to be executed, which includes test conditions and expectations.
//
// Returns an error if the test fails due to an error in execution or if the test conditions are not met.
// If the test passes, including successful execution and meeting of the test conditions, the function
// returns nil. The function no longer logs fatal errors but returns them to the caller for handling.
func RunTimelineTests(ctx context.Context, buildClient BuildClient, projectName string, buildID *int, test TimelineTest) error {
	var pass bool
	var err error

	pass, err = GetBuildStageStatus(ctx, buildClient, projectName, buildID, test)
	if err != nil {
		return fmt.Errorf("test failed for %s: %v", test.Name, err)
	}

	if !pass {
		return fmt.Errorf("test failed for %s: condition not met", test.Name)
	}

	fmt.Printf("Test passed for %s\n", test.Name)
	return nil
}

// GetTestsDefinition reads a YAML file from a specified path and unmarshalls it into slices of BuildTest and TimelineTest.
// This function is used for parsing test definitions from a YAML configuration file, allowing for dynamic test specifications.
//
// Parameters:
// filePath - The path to the YAML file that contains the test definitions.
//
// Returns two slices: one of BuildTest and another of TimelineTest, and an error. Each slice contains the respective test
// definitions extracted from the YAML file. The BuildTest slice contains tests related to build processes, whereas the
// TimelineTest slice contains tests that pertain to timeline events in a build.
//
// In case of errors in reading the file or unmarshalling the content, the function returns an error with a detailed
// description of the issue. This allows the caller to decide how to handle such scenarios, instead of terminating
// the execution immediately. This change in design provides more flexibility in error handling.
func GetTestsDefinition(filePath string) (buildTests []BuildTest, timelineTests []TimelineTest, err error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading tests file: %v", err)
	}

	var tests Tests
	err = yaml.Unmarshal(fileContent, &tests)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling tests: %v", err)
	}

	return tests.BuildTests, tests.TimelineTests, nil
}

// NewRunPipelineArgs is a function that creates the arguments required to trigger run an Azure DevOps (ADO) pipeline.
// It takes a map of pipeline parameters, an ADO configuration,
// and a variadic slice of pipeline run arguments options as input parameters.
//
// Parameters:
// templateParameters - A map of key-value pairs that represent the parameters for the pipeline.
// adoConfig - The configuration for the ADO pipelines organization.
// pipelineRunArgs - A variadic slice of options for the pipeline run arguments.
//
//	Each option is a function that modifies the pipeline run arguments.
//
// If all options are successfully applied, the function returns the run arguments and nil for the error.
func NewRunPipelineArgs(templateParameters map[string]string, adoConfig Config, pipelineRunArgs ...RunPipelineArgsOptions) (pipelines.RunPipelineArgs, error) {
	pipelineID := &adoConfig.ADOPipelineID

	// Create the pipeline run arguments.
	adoRunPipelineArgs := pipelines.RunPipelineArgs{
		Project:    &adoConfig.ADOProjectName,
		PipelineId: pipelineID,
		RunParameters: &pipelines.RunPipelineParameters{
			PreviewRun:         ptr.To(false),
			TemplateParameters: &templateParameters,
		},
	}
	// Set pipeline version if it's not 0 in the global ADO configuration.
	if adoConfig.ADOPipelineVersion != 0 {
		adoRunPipelineArgs.PipelineVersion = &adoConfig.ADOPipelineVersion
	}
	// Apply the pipeline run arguments options.
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

// RunPipelineArgsOptions is a type that defines a function that modifies the RunPipelineArgs.
// This function takes a pointer to a RunPipelineArgs and returns an error.
// It is used to pass functional options to the NewRunPipelineArgs function.
type RunPipelineArgsOptions func(*pipelines.RunPipelineArgs) error

// PipelinePreviewRun is a function that returns a RunPipelineArgsOptions.
// This function sets the PreviewRun field of the RunParameters to true and
// reads the override YAML file with ADO pipeline definition from the provided path to set the YamlOverride field.
//
// Parameters:
// overrideYamlPath - The path to the ADO pipeline definition YAML file.
//
// Returns a function that modifies the RunPipelineArgs.
// This function reads the override YAML file, converts it to a string,
// and sets the YamlOverride field of the RunParameters.
func PipelinePreviewRun(overrideYamlPath string) RunPipelineArgsOptions {
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
