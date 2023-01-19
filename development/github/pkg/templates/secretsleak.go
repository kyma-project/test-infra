package templates

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"github.com/kyma-project/test-infra/development/types"
)

type IssueData interface {
	RenderBody (*bytes.Buffer, error)
}

type SecretsLeakIssueData struct {
	pubsub.ProwMessage
	types.SecretsLeakScannerMessage
	SecretsLeaksScannerID string
}

func (s *SecretsLeakIssueData) RenderBody() (*bytes.Buffer, error) {
	templateContent := "Found secret in {{.JobName}} {{.JobType}} prowjob logs. Moved logs to secure location. You can see logs [here]({{.GcsPath}})\n\n" +
		"Please check leaked secret, rotate it and clean git history if it was committed. Please remove code and settings which print secret in logs.\n\n" +
		"Detected leaks.\n\n{{range .LeaksReport}}- Description: {{.Description}}\n  StartLine: {{.StartLine}}\n  EndLine: {{.EndLine}}\n  " +
		"StartColumn: {{.StartColumn}}\n  EndColumn: {{.EndColumn}}\n  File: {{.File}}\n  RuleID: {{.RuleID}}\n{{end}}\n\n" +
		"<!--\nDO NOT REMOVE THIS COMMENT\nSECRETS LEAK SCANNER METADATA\nsecretsleakscanner_id={{.SecretsLeaksScannerID}}\nsecretsleakscanner_jobname={{.JobName}}\nsecretsleakscanner_jobtype={{.JobType}}\n--!>"
	tmpl, err := template.New("secretsLeakFoundIssueBody").Parse(templateContent)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, s)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
