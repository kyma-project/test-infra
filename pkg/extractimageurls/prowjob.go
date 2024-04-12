package extractimageurls

import (
	"context"
	"fmt"
	"github.com/kyma-project/test-infra/pkg/github/client"

	"github.com/google/go-github/v48/github"
	"sigs.k8s.io/prow/prow/config"
	"sigs.k8s.io/yaml"
)

// Repository contains information about github repository used by FromInRepoConfig
type Repository struct {
	// Name is repository name
	Name string

	// Owner is a owner of the repository
	Owner string
}

func (repo *Repository) String() string {
	return fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
}

// FromProwJobConfig parses JobConfig from prow library and returns a slice of image urls
func FromProwJobConfig(config config.JobConfig) []string {
	images := []string{}
	images = append(images, extractImageUrlsFromPeriodicsProwJobs(config.Periodics)...)
	images = append(images, extractImageUrlsFromPresubmitsProwJobs(config.PresubmitsStatic)...)
	images = append(images, extractImageUrlsFromPostsubmitsJobs(config.PostsubmitsStatic)...)

	return images
}

// FromInRepoConfig fetch prowjobs from .prow directory or .prow.yaml file
// on specified gihub repository and main branch, then extract image urls from them
func FromInRepoConfig(repository Repository, ghToken string) ([]string, error) {
	// Create github authenticated client
	ctx := context.Background()
	ghClient, err := client.NewClient(ctx, ghToken)
	if err != nil {
		return nil, err
	}

	// fetch files from .prow, failback to .prow.yaml if not exists
	repositoryContent, directoryContent, resp, err := ghClient.Repositories.GetContents(ctx, repository.Owner, repository.Name, ".prow", &github.RepositoryContentGetOptions{})
	if resp.StatusCode == 404 {
		repositoryContent, _, resp, err = ghClient.Repositories.GetContents(ctx, repository.Owner, repository.Name, ".prow.yaml", &github.RepositoryContentGetOptions{})
	}
	if err != nil {
		return nil, err
	}

	if ok, err := client.IsStatusOK(resp); !ok {
		return nil, err
	}

	var images []string
	if directoryContent != nil {
		for _, fileContent := range directoryContent {
			imgs, err := extractImageUrlsFromRepositoryContent(fileContent)
			if err != nil {
				return nil, err
			}

			images = append(images, imgs...)
		}
	} else {
		imgs, err := extractImageUrlsFromRepositoryContent(repositoryContent)
		if err != nil {
			return nil, err
		}

		images = append(images, imgs...)
	}

	return images, nil
}

// extractImageUrlsFromRepositoryContent return image urls from given repository content
func extractImageUrlsFromRepositoryContent(content *github.RepositoryContent) ([]string, error) {
	file, err := content.GetContent()
	if err != nil {
		return nil, err
	}

	var cfg config.JobConfig
	err = yaml.Unmarshal([]byte(file), &cfg)
	if err != nil {
		return nil, err
	}

	var images []string
	images = append(images, extractImageUrlsFromPeriodicsProwJobs(cfg.Periodics)...)
	images = append(images, extractImageUrlsFromPostsubmitsJobs(cfg.PostsubmitsStatic)...)
	images = append(images, extractImageUrlsFromPresubmitsProwJobs(cfg.PresubmitsStatic)...)

	return images, nil
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
			if job.Spec == nil {
				continue
			}
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
			if job.Spec == nil {
				continue
			}
			for _, container := range job.Spec.Containers {
				images = append(images, container.Image)
			}
		}
	}

	return images
}
