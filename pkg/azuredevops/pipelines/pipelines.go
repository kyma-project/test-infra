// Package pipelines provides a clients for calling Azure DevOps pipelines API
// TODO: Add more structured logging with debug severity to track execution in case of troubleshooting
package pipelines

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	adov7 "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"
	"golang.org/x/net/context"
	"k8s.io/utils/ptr"
)

type Client interface {
	RunPipeline(ctx context.Context, args pipelines.RunPipelineArgs) (*pipelines.Run, error)
	GetRun(ctx context.Context, args pipelines.GetRunArgs) (*pipelines.Run, error)
}

type BuildClient interface {
	GetBuildLogs(ctx context.Context, args build.GetBuildLogsArgs) (*[]build.BuildLog, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
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
