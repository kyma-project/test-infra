package detectors

import (
	"k8s.io/test-infra/prow/config"
)

func Extract(path string) ([]string, error) {
	cfg, err := config.ReadJobConfig(path)
	if err != nil {
		return nil, err
	}

	images := extract(cfg)

	return images, nil
}

func ExtractFromJobConfig(config config.JobConfig) []string {
	return extract(config)
}

func extract(config config.JobConfig) []string {
	images := []string{}
	images = append(images, extractPeriodics(config.Periodics)...)
	images = append(images, extractPresubmits(config.PresubmitsStatic)...)
	images = append(images, extractPostsubmits(config.PostsubmitsStatic)...)

	return images
}

func extractPeriodics(periodics []config.Periodic) []string {
	var images []string
	for _, job := range periodics {
		for _, container := range job.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images
}

func extractPresubmits(presubmits map[string][]config.Presubmit) []string {
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

func extractPostsubmits(postsubmits map[string][]config.Postsubmit) []string {
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
