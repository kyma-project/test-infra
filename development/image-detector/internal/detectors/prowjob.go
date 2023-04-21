package detectors

import (
	"k8s.io/test-infra/prow/config"
)

type ProwJob struct{}

func NewProwJobDetector() *ProwJob {
	return &ProwJob{}
}

func (p *ProwJob) Check(path string) bool {
	_, err := config.ReadJobConfig(path)
	return err == nil
}

func (p *ProwJob) Extract(path string) ([]string, error) {
	cfg, err := config.ReadJobConfig(path)
	if err != nil {
		return nil, err
	}

	images := p.extract(cfg)

	return images, nil
}

func (p *ProwJob) ExtractFromJobConfig(config config.JobConfig) []string {
	return p.extract(config)
}

func (p *ProwJob) extract(config config.JobConfig) []string {
	images := []string{}
	images = append(images, p.extractPeriodics(config.Periodics)...)
	images = append(images, p.extractPresubmits(config.PresubmitsStatic)...)
	images = append(images, p.extractPostsubmits(config.PostsubmitsStatic)...)

	return images
}

func (p *ProwJob) extractPeriodics(periodics []config.Periodic) []string {
	var images []string
	for _, job := range periodics {
		for _, container := range job.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images
}

func (p *ProwJob) extractPresubmits(presubmits map[string][]config.Presubmit) []string {
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

func (p *ProwJob) extractPostsubmits(postsubmits map[string][]config.Postsubmit) []string {
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
