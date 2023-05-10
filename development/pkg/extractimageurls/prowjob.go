package extractimageurls

import (
	"k8s.io/test-infra/prow/config"
)

// FromProwJobConfigFile find images from prow job config file and returns it as slice
func FromProwJobConfigFile(path string) ([]string, error) {
	cfg, err := config.ReadJobConfig(path)
	if err != nil {
		return nil, err
	}

	images := extractImagesFromProwJobConfig(cfg)

	return images, nil
}

// FromProwJobConfig parses JobConfig from prow library and returns it as slice
func FromProwJobConfig(config config.JobConfig) []string {
	return extractImagesFromProwJobConfig(config)
}

func extractImagesFromProwJobConfig(config config.JobConfig) []string {
	images := []string{}
	images = append(images, extractImagesFromPeriodicsProwJobs(config.Periodics)...)
	images = append(images, extractImagesFromPresubmitsProwJobs(config.PresubmitsStatic)...)
	images = append(images, extractImagesFromPostsubmitsJobs(config.PostsubmitsStatic)...)

	return images
}

func extractImagesFromPeriodicsProwJobs(periodics []config.Periodic) []string {
	var images []string
	for _, job := range periodics {
		for _, container := range job.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images
}

func extractImagesFromPresubmitsProwJobs(presubmits map[string][]config.Presubmit) []string {
	var images []string
	for _, repo := range presubmits {
		for _, job := range repo {
			for _, container := range job.Spec.Containers {
				images = append(images, container.Image)
			}
		}
	}

	return images
}

func extractImagesFromPostsubmitsJobs(postsubmits map[string][]config.Postsubmit) []string {
	var images []string
	for _, repo := range postsubmits {
		for _, job := range repo {
			for _, container := range job.Spec.Containers {
				images = append(images, container.Image)
			}
		}
	}

	return images
}
