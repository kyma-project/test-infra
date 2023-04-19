package detectors

import (
	"k8s.io/test-infra/prow/config"
)

func Check(path string) bool {
	_, err := config.ReadJobConfig(path)
	return err == nil
}

func Extract(path string) ([]string, error) {
	cfg, err := config.ReadJobConfig(path)
	if err != nil {
		return nil, err
	}

	images := extract(cfg)

	return images, nil
}

func extract(config config.JobConfig) []string {
	images := []string{}
	images = appendIfMissing(images, extractPeriodics(config.Periodics)...)
	images = appendIfMissing(images, extractPresubmits(config.PresubmitsStatic)...)
	images = appendIfMissing(images, extractPostsubmits(config.PostsubmitsStatic)...)

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

func appendIfMissing[T comparable](slice []T, elements ...T) []T {
	cache := make(map[T]T)

	for _, el := range slice {
		cache[el] = el
	}

	for _, el := range elements {
		if _, ok := cache[el]; !ok {
			slice = append(slice, el)
		}
	}

	return slice
}
