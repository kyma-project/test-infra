package gcrcleaner

import (
	"context"
	"sort"

	"github.com/google/go-containerregistry/pkg/authn"
	gcrname "github.com/google/go-containerregistry/pkg/name"
	gcrgoogle "github.com/google/go-containerregistry/pkg/v1/google"
	gcrremote "github.com/google/go-containerregistry/pkg/v1/remote"
)

// TODO better name
type RepoAPIWrapper struct {
	Context context.Context
	Auth    authn.Authenticator
}

type ImageAPIWrapper struct {
	Context context.Context
	Auth    authn.Authenticator
	// repository string
}

func (raw *RepoAPIWrapper) ListSubrepositories(repoName string) ([]string, error) {

	repo, err := gcrname.NewRepository(repoName)
	if err != nil {
		return nil, err
	}

	childRepos, err := gcrremote.Catalog(raw.Context, repo.Registry, gcrremote.WithAuth(raw.Auth))
	if err != nil {
		return nil, err
	}

	sort.Strings(childRepos)
	return childRepos, nil
}

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

func (iw *ImageAPIWrapper) DeleteImage(registry, repoName string, imageSHA string) error {
	return nil
}
