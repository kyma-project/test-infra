package detectors

import (
	"k8s.io/test-infra/prow/config"
)

type ProwJob struct {
}

func (p *ProwJob) Check(path string) bool {
	_, err := config.ReadJobConfig(path)
	return err == nil
}

func (p *ProwJob) Extract(path string) ([]string, error) {
	config, err := config.ReadJobConfig(path)
	if err != nil {
		return nil, err
	}

	images := extract(config)

	return images, nil
}

func extract(config config.JobConfig) []string {
	var images []string
	images = append(images, extractPeriodics(config.Periodics)...)
	images = append(images, extractPresubmits(config.PresubmitsStatic)...)

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
