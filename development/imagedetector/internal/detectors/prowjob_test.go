package detectors

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/test-infra/prow/config"
)

func TestExtract(t *testing.T) {
	tc := []struct {
		name           string
		jobConfig      config.JobConfig
		expectedImages []string
	}{
		{
			name: "basic periodic prowjob, pass",
			jobConfig: config.JobConfig{
				Periodics: []config.Periodic{
					{
						JobBase: config.JobBase{
							Spec: &v1.PodSpec{
								Containers: []v1.Container{
									{Image: "nginx:1.14.2"},
								},
							},
						},
					},
				},
			},
			expectedImages: []string{"nginx:1.14.2"},
		},
		{
			name: "basic prowjobs file, pass",
			jobConfig: config.JobConfig{
				Periodics: []config.Periodic{
					{
						JobBase: config.JobBase{
							Spec: &v1.PodSpec{
								Containers: []v1.Container{
									{Image: "nginx:1.14.2"},
								},
							},
						},
					},
				},
				PresubmitsStatic: map[string][]config.Presubmit{
					"test-org/test-repo": {
						{
							JobBase: config.JobBase{
								Spec: &v1.PodSpec{
									Containers: []v1.Container{
										{Image: "test.com/test-org/test-repo/image:test-tag"},
									},
								},
							},
						},
					},
				},
			},
			expectedImages: []string{"nginx:1.14.2", "test.com/test-org/test-repo/image:test-tag"},
		},
		{
			name: "postsubmits prowjob, pass",
			jobConfig: config.JobConfig{
				PostsubmitsStatic: map[string][]config.Postsubmit{
					"test-org/test-repo": {
						{
							JobBase: config.JobBase{
								Spec: &v1.PodSpec{
									Containers: []v1.Container{
										{Image: "test.com/test-org/test-repo/image:test-tag"},
									},
								},
							},
						},
					},
				},
			},
			expectedImages: []string{"test.com/test-org/test-repo/image:test-tag"},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			images := extract(c.jobConfig)

			if !reflect.DeepEqual(images, c.expectedImages) {
				t.Errorf("%v != %v", images, c.expectedImages)
			}
		})
	}
}
