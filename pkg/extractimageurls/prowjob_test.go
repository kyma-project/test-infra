package extractimageurls

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/prow/prow/config"
)

func TestFromProwJobConfig(t *testing.T) {
	tc := []struct {
		name           string
		expectedImages []string
		jobConfig      config.JobConfig
	}{
		{
			name:           "basic periodic prowjob, pass",
			expectedImages: []string{"nginx:1.14.2"},
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
		},
		{
			name:           "basic prowjobs file, pass",
			expectedImages: []string{"nginx:1.14.2", "test.com/test-org/test-repo/image:test-tag"},
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
		},
		{
			name:           "postsubmits prowjob, pass",
			expectedImages: []string{"test.com/test-org/test-repo/image:test-tag"},
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
		},
		{
			name:           "empty prowjob, pass no images",
			jobConfig:      config.JobConfig{},
			expectedImages: []string{},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			images := FromProwJobConfig(c.jobConfig)

			if !reflect.DeepEqual(images, c.expectedImages) {
				t.Errorf("ExtractImagesFromProwJobs(): Got %v, but expected %v", images, c.expectedImages)
			}
		})
	}
}
