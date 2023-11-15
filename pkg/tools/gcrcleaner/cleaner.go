package gcrcleaner

import (
	"regexp"
	"time"

	gcrname "github.com/google/go-containerregistry/pkg/name"
	gcrgoogle "github.com/google/go-containerregistry/pkg/v1/google"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery --name=RepoAPI --output=automock --outpkg=automock --case=underscore

// RepoAPI abstracts access to Docker repos
type RepoAPI interface {
	ListSubrepositories(repoName string) ([]string, error)
}

//go:generate mockery --name=ImageAPI --output=automock --outpkg=automock --case=underscore

// ImageAPI abstracts access to Docker images
type ImageAPI interface {
	ListImages(registry, repoName string) (map[string]gcrgoogle.ManifestInfo, error)
	DeleteImage(registry, repoName string, digest string, manifest gcrgoogle.ManifestInfo) error
}

// GCRCleaner deletes Docker images created by prow jobs.
type GCRCleaner struct {
	repoAPI           RepoAPI
	imageAPI          ImageAPI
	shouldRemoveRepo  RepoRemovalPredicate
	shouldRemoveImage ImageRemovalPredicate
}

// New returns a new instance of GCRCleaner
func New(repoAPI RepoAPI, imageAPI ImageAPI, shouldRemoveRepo RepoRemovalPredicate, shouldRemoveImage ImageRemovalPredicate) *GCRCleaner {
	return &GCRCleaner{repoAPI, imageAPI, shouldRemoveRepo, shouldRemoveImage}
}

// Run executes image removal process for specified Docker repository
func (gcrc *GCRCleaner) Run(repoName string, makeChanges bool) (allSucceeded bool, err error) {
	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}
	allSucceeded = true

	repo, err := gcrname.NewRepository(repoName)
	if err != nil {
		return false, err
	}
	registry := repo.Registry.RegistryStr()

	repos, err := gcrc.repoAPI.ListSubrepositories(repoName)
	if err != nil {
		return false, err
	}

	for _, repo := range repos {
		if gcrc.shouldRemoveRepo(repo) {
			log.Infof("Cleaning images in %s/%s repository", registry, repo)

			// get all images for each repo
			images, err := gcrc.imageAPI.ListImages(registry, repo)
			if err != nil {
				return false, err
			}
			for sha, image := range images {

				// check if the image is a candidate for deletion
				if gcrc.shouldRemoveImage(&image) {
					if makeChanges {
						err = gcrc.imageAPI.DeleteImage(registry, repo, sha, image)
					}
					if err != nil {
						log.Errorf("deleting image %s/%s@%s: %#v", registry, repo, sha, err)
						allSucceeded = false
					} else {
						log.Infof("%sImage %s/%s@%s deleted", msgPrefix, registry, repo, sha)
					}

				}
			}
		}
	}
	return allSucceeded, nil
}

// RepoRemovalPredicate returns true when images in repo should be considered for deletion (name matches removal criteria)
type RepoRemovalPredicate func(string) bool

// ImageRemovalPredicate returns true when image should be deleted (matches removal criteria)
type ImageRemovalPredicate func(*gcrgoogle.ManifestInfo) bool

// NewRepoFilter is a default RepoRemovalPredicate factory
// Repo is matching the criteria if it's:
// - Name does not match gcrNameIgnoreRegex
func NewRepoFilter(gcrNameIgnoreRegex *regexp.Regexp) RepoRemovalPredicate {
	return func(repo string) bool {
		nameMatches := false
		if gcrNameIgnoreRegex.String() != "" {
			nameMatches = gcrNameIgnoreRegex.MatchString(repo)
		}
		return !nameMatches

	}
}

// NewImageFilter is a default ImageRemovalPredicate factory
// Image is matching the criteria if it's:
// - CreationTimestamp indicates that it is created more than ageInHours ago.
func NewImageFilter(ageInHours int) ImageRemovalPredicate {
	return func(image *gcrgoogle.ManifestInfo) bool {
		oldEnough := false

		imageAgeHours := time.Since(image.Created).Hours() - float64(ageInHours)
		oldEnough = imageAgeHours > 0

		return oldEnough
	}
}
