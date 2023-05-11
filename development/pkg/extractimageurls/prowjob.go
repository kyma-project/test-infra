package extractimageurls

import (
	"k8s.io/test-infra/prow/config"
)

// FromProwJobConfig parses JobConfig from prow library and returns a slice of image urls
func FromProwJobConfig(config config.JobConfig) []string {
	images := []string{}
	images = append(images, extractImageUrlsFromPeriodicsProwJobs(config.Periodics)...)
	images = append(images, extractImageUrlsFromPresubmitsProwJobs(config.PresubmitsStatic)...)
	images = append(images, extractImageUrlsFromPostsubmitsJobs(config.PostsubmitsStatic)...)

	return images
}

// extractImageUrlsFromPeriodicsProwJobs returns slice of image urls from given periodic prow jobs
func extractImageUrlsFromPeriodicsProwJobs(periodics []config.Periodic) []string {
	var images []string
	for _, job := range periodics {
		for _, container := range job.Spec.Containers {
			images = append(images, container.Image)
		}
	}

	return images
}

// extractImageUrlsFromPresubmitsProwJobs returns slice of image urls from given presubmit prow jobs
func extractImageUrlsFromPresubmitsProwJobs(presubmits map[string][]config.Presubmit) []string {
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

// extractImageUrlsFromPostsubmitProwJobs returns slice of image urls from given postsubmit prow jobs
func extractImageUrlsFromPostsubmitsJobs(postsubmits map[string][]config.Postsubmit) []string {
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
