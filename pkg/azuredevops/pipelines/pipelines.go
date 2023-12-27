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
