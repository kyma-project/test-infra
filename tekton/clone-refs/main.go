package main

import (
	"encoding/json"
	"os"

	v1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/clonerefs"
	"k8s.io/test-infra/prow/logrusutil"
	// "k8s.io/test-infra/prow/pod-utils/options"

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

	jobSpecEnv := os.Getenv("JOB_SPEC")

	jobSpec := map[string]interface{}{}
	err := json.Unmarshal([]byte(jobSpecEnv), &jobSpec)
	if err != nil {
		logrus.Fatalf("Could not unmarshal job spec: %v", err)
	}

	if val, ok := jobSpec["refs"]; !ok {
		refsString, _ = json.Marshal(val)
		err = json.Unmarshal(refsString, &refs)
		if err != nil {
			logrus.Fatalf("Could not unmarshal refs: %v", err)
		}
		gitRefs = append(gitRefs, refs)
	}

	if val, ok := jobSpec["extra_refs"]; ok {
		extraRefsString, _ = json.Marshal(val)
		err = json.Unmarshal(extraRefsString, &extraRefs)
		if err != nil {
			logrus.Fatalf("Could not unmarshal extra refs: %v", err)
		}
		gitRefs = append(gitRefs, extraRefs...)
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
