package extractimageurls

import (
	"reflect"
	"strings"
	"testing"
)

func TestFromTektonTask(t *testing.T) {
	tc := []struct {
		Name           string
		ExpectedImages []string
		WantErr        bool
		FileContent    string
	}{
		{
			Name:           "simple task definition, pass",
			ExpectedImages: []string{"gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:v0.40.2"},
			WantErr:        false,
			FileContent: `apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: some-task
  labels:
    app.kubernetes.io/version: "0.1"
  annotations:
    tekton.dev/pipelines.minVersion: "0.36.0"
    tekton.dev/categories: Git
    tekton.dev/tags: git
    tekton.dev/displayName: "some-task"
    tekton.dev/platforms: "linux/amd64"
spec:
  description: some description
  workspaces:
    - name: logs
      description: "workspace description"
      mountPath: /path/to/logs
      optional: true
  steps:
    - name: clone
      image: gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:v0.40.2`,
		},
		{
			Name:           "image url in param default, pass",
			ExpectedImages: []string{"gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:v0.40.2"},
			WantErr:        false,
			FileContent: `apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: some-task
  labels:
    app.kubernetes.io/version: "0.1"
  annotations:
    tekton.dev/pipelines.minVersion: "0.36.0"
    tekton.dev/categories: Git
    tekton.dev/tags: git
    tekton.dev/displayName: "some-task"
    tekton.dev/platforms: "linux/amd64"
spec:
  description: some description
  workspaces:
    - name: logs
      description: "workspace description"
      mountPath: /path/to/logs
      optional: true
  params:
    - name: gitInitImage
      description: The image providing the git-init binary that this Task runs.
      type: string
      default: "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:v0.40.2"
  steps:
    - name: clone
      image: "$(params.gitInitImage)"`,
		},
		{
			Name:           "image url in param, no default value, pass with empty images",
			ExpectedImages: nil,
			WantErr:        false,
			FileContent: `apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: some-task
  labels:
    app.kubernetes.io/version: "0.1"
  annotations:
    tekton.dev/pipelines.minVersion: "0.36.0"
    tekton.dev/categories: Git
    tekton.dev/tags: git
    tekton.dev/displayName: "some-task"
    tekton.dev/platforms: "linux/amd64"
spec:
  description: some description
  workspaces:
    - name: logs
      description: "workspace description"
      mountPath: /path/to/logs
      optional: true
  params:
    - name: gitInitImage
      description: The image providing the git-init binary that this Task runs.
      type: string
  steps:
    - name: clone
      image: "$(params.gitInitImage)"`,
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			images, err := FromTektonTask(strings.NewReader(c.FileContent))
			if err != nil && !c.WantErr {
				t.Errorf("error occurred but not expected: %s", err)
			}

			if !reflect.DeepEqual(images, c.ExpectedImages) {
				t.Errorf("FromTektonTask(): Got %v, but expected %v", images, c.ExpectedImages)
			}
		})
	}
}
