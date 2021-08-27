package gcrcleaner

import (
	"context"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	gcrname "github.com/google/go-containerregistry/pkg/name"
	gcrgoogle "github.com/google/go-containerregistry/pkg/v1/google"
	gcrremote "github.com/google/go-containerregistry/pkg/v1/remote"
	log "github.com/sirupsen/logrus"
)

// RepoAPIWrapper abstracts Docker repo API
type RepoAPIWrapper struct {
	Context context.Context
	Auth    authn.Authenticator
}

// ImageAPIWrapper abstracts Docker image API
type ImageAPIWrapper struct {
	Context context.Context
	Auth    authn.Authenticator
	// repository string
}

// ListSubrepositories implements RepoAPIW.ListSubrepositories function
func (raw *RepoAPIWrapper) ListSubrepositories(repoName string) ([]string, error) {
	var repositories []string
	repo, err := gcrname.NewRepository(repoName)
	if err != nil {
		return nil, err
	}

	childRepos, err := gcrremote.Catalog(raw.Context, repo.Registry, gcrremote.WithAuth(raw.Auth))
	if err != nil {
		return nil, err
	}

	//add only repos that start with wanted repo name
	for _, childRepo := range childRepos {
		if strings.HasPrefix(childRepo, repo.RepositoryStr()) {
			repositories = append(repositories, childRepo)
		}
	}
	sort.Strings(repositories)
	return repositories, nil
}

// ListImages implements ImageAPI.ListImages function
func (iw *ImageAPIWrapper) ListImages(registry, repoName string) (map[string]gcrgoogle.ManifestInfo, error) {
	repo, err := gcrname.NewRepository(registry + "/" + repoName)
	if err != nil {
		return nil, err
	}

	tags, err := gcrgoogle.List(repo, gcrgoogle.WithAuth(iw.Auth))
	if err != nil {
		return nil, err
	}

	// return map of manifests, each key is in "sha256:XXXXXXXXXXX" format
	return tags.Manifests, nil
}

// DeleteImage implements ImageAPI.DeleteImage function
func (iw *ImageAPIWrapper) DeleteImage(registry, repoName string, digest string, manifest gcrgoogle.ManifestInfo) error {
	repo, err := gcrname.NewRepository(registry + "/" + repoName)
	if err != nil {
		return err
	}

	// delete all tags
	for _, tag := range manifest.Tags {
		taggedName := repo.Tag(tag)
		log.Info(taggedName)
		err := gcrremote.Delete(taggedName, gcrremote.WithAuth(iw.Auth))
		return err
	}

	// delete image
	reference := repo.Digest(digest)
	log.Info(reference)
	err = gcrremote.Delete(reference, gcrremote.WithAuth(iw.Auth))
	return err
}
