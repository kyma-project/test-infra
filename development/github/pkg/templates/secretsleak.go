package templates

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"github.com/kyma-project/test-infra/development/types"
)

type IssueData interface {
	RenderBody(*bytes.Buffer, error)
}

type SecretsLeakIssueData struct {
	pubsub.ProwMessage
	types.SecretsLeakScannerMessage
	SecretsLeaksScannerID      string
	KymaSecurityGithubTeamName string // kyma/security
}

func (s *SecretsLeakIssueData) RenderBody() (*bytes.Buffer, error) {
	templateContent := `
Found secret in {{.JobName}} {{.JobType}} prowjob logs. Moved logs to secure location. You can see logs [here]({{.GcsPath}})


Please check leaked secret, rotate it and clean git history if it was committed. Please remove code and settings which print secret in logs.
Detected leaks.
{{- range .LeaksReport}}
- Description: {{.Description}}
  StartLine: {{.StartLine}}
  EndLine: {{.EndLine}}
  StartColumn: {{.StartColumn}}
  EndColumn: {{.EndColumn}}
  File: {{.File}}
  RuleID: {{.RuleID}}
{{- end}}

@{{.KymaSecurityGithubTeamName}} please note we got this.

<!--
DO NOT REMOVE THIS COMMENT
SECRETS LEAK SCANNER METADATA
secretsleakscanner_id={{.SecretsLeaksScannerID}}
secretsleakscanner_jobname={{.JobName}}
secretsleakscanner_jobtype={{.JobType}}
--!>`

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
