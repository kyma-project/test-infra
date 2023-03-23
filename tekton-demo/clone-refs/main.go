// This is a usage of clonerefs tool from kubernetes/test-infra repo adjusted to work with Tekton as a Task.
// It is used to clone git repositories specified by refs and extra-refs fields in JOB_SPEC env var.
// JOB_SPEC env var is provided by Prow as pipeline param and contains information about the job that is currently running.
// It is supposed to be used as a replacement for piepline git resource in a Tekton Pipeline Task that is triggered by Prow.
package main

import (
	"encoding/json"
	"os"

	v1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/clonerefs"
	"k8s.io/test-infra/prow/logrusutil"

	"github.com/sirupsen/logrus"
)

func main() {
	logrusutil.ComponentInit()

	var (
		refsString      []byte
		extraRefsString []byte
		refs            v1.Refs
		extraRefs       []v1.Refs
		gitRefs         []v1.Refs
	)

	// Check if /logs directory exists. If not, create it.
	// This makes logs location consistent with prowjobs running on kubernetes agent without requiring a dedicated workspace.
	if _, err := os.Stat("/logs"); os.IsNotExist(err) {
		err = os.Mkdir("/logs", 0755)
		if err != nil {
			logrus.Fatalf("Could not create /logs directory: %v", err)
		}
	}

	// JOB_SPEC env var is provided by the Prow.
	jobSpecEnv := os.Getenv("JOB_SPEC")

	jobSpec := map[string]interface{}{}
	err := json.Unmarshal([]byte(jobSpecEnv), &jobSpec)
	if err != nil {
		logrus.Fatalf("Could not unmarshal job spec: %v", err)
	}

	// Extract refs from JOB_SPEC.
	if val, ok := jobSpec["refs"]; ok {
		refsString, _ = json.Marshal(val)
		err = json.Unmarshal(refsString, &refs)
		if err != nil {
			logrus.Fatalf("Could not unmarshal refs: %v", err)
		}
		gitRefs = append(gitRefs, refs)
		logrus.Infof("gitRefs: %v\n", gitRefs)
	}

	// Extract extra_refs from JOB_SPEC.
	if val, ok := jobSpec["extra_refs"]; ok {
		extraRefsString, _ = json.Marshal(val)
		err = json.Unmarshal(extraRefsString, &extraRefs)
		if err != nil {
			logrus.Fatalf("Could not unmarshal extra refs: %v", err)
		}
		gitRefs = append(gitRefs, extraRefs...)
		logrus.Infof("gitRefs: %v\n", gitRefs)
	}

	o := clonerefs.Options{
		SrcRoot:            os.Getenv("SRC_ROOT"),
		Log:                os.Getenv("LOG"),
		GitUserName:        clonerefs.DefaultGitUserName,
		GitUserEmail:       clonerefs.DefaultGitUserEmail,
		GitRefs:            gitRefs,
		Fail:               true,
		GitHubAPIEndpoints: []string{"https://api.github.com"},
	}

	if err := o.Validate(); err != nil {
		logrus.Fatalf("Invalid options: %v", err)
	}

	if err := o.Run(); err != nil {
		logrus.WithError(err).Fatal("Failed to clone refs")
	}

	logrus.Info("Finished cloning refs")
}
