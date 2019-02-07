package slack

import (
	"fmt"
	"k8s.io/test-infra/prow/apis/prowjobs/v1"
	"testing"
)

func Test(t *testing.T) {
	pj := v1.ProwJob{
		Spec: v1.ProwJobSpec{
			Job: "pre-master-kyma-integration",
		},
		Status: v1.ProwJobStatus{
			State: v1.FailureState,
			URL:   "https://status.build.kyma-project.io/view/gcs/kyma-prow-logs/pr-logs/pull/kyma-project_kyma/2514/pre-master-kyma-integration/1090592915127799810",
		},
	}
	fmt.Printf("ProwJob %s *FAILED:* State <%s>. Click <%s|here> to see the Job details. ", pj.Spec.Job, pj.Status.State, pj.Status.URL)
}
