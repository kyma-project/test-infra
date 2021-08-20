package gcrcleaner

import (
	"errors"
	"regexp"
	"time"

	gcrauth "github.com/google/go-containerregistry/pkg/authn"
	gcrname "github.com/google/go-containerregistry/pkg/name"
	gcrgoogle "github.com/google/go-containerregistry/pkg/v1/google"
	log "github.com/sirupsen/logrus"
)

//go:generate mockery --name=RepoAPI --output=automock --outpkg=automock --case=underscore

// aaa
type RepoAPI interface {
	ListSubrepositories(repoName string) ([]string, error)
}

//go:generate mockery --name=Image --output=automock --outpkg=automock --case=underscore

// aaa
type ImageAPI interface {
	ListImages(registry, repoName string) (map[string]gcrgoogle.ManifestInfo, error)
	DeleteImage(registry, repoName string, imageSHA string) error
}

type GCRCleaner struct {
	auth              gcrauth.Authenticator
	repoAPI           RepoAPI
	imageAPI          ImageAPI
	shouldRemoveRepo  RepoRemovalPredicate
	shouldRemoveImage ImageRemovalPredicate
}

func New(auth gcrauth.Authenticator, repoAPI RepoAPI, imageAPI ImageAPI, shouldRemoveRepo RepoRemovalPredicate, shouldRemoveImage ImageRemovalPredicate) *GCRCleaner {
	return &GCRCleaner{auth, repoAPI, imageAPI, shouldRemoveRepo, shouldRemoveImage}
}

func (gcrc *GCRCleaner) Run(repoName string, makeChanges bool) (allSucceeded bool, err error) {
	var msgPrefix string
	if !makeChanges {
		msgPrefix = "[DRY RUN] "
	}

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
		shouldRemoveRepo, err := gcrc.shouldRemoveRepo(repo)
		if err != nil {

		} else if shouldRemoveRepo {
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
						err = gcrc.imageAPI.DeleteImage(registry, repo, sha)
					}
					if err != nil {
						log.Errorf("deleting image %s/%s@%s: %#v", registry, repo, sha, err)
					} else {
						log.Infof("%s Image %s/%s@%s deleted", msgPrefix, registry, repo, sha)
					}

				}
			}
		} else {
			log.Infof("%s won't be considered for deletion", repo)
		}
	}
	return false, errors.New("run not implemented")
}

// RepoRemovalPredicate returns true when images in repo should be considered for deletion (matches removal criteria)
type RepoRemovalPredicate func(string) (bool, error)

// ImageRemovalPredicate returns true when image should be deleted (matches removal criteria)
type ImageRemovalPredicate func(*gcrgoogle.ManifestInfo) bool

// NewRepoFilter is a default IPRemovalPredicate factory
// Repo is matching the criteria if it's:
// - Name does not match gcrNameIgnoreRegex
func NewRepoFilter(gcrNameIgnoreRegex *regexp.Regexp) RepoRemovalPredicate {
	return func(repo string) (bool, error) {
		nameMatches := gcrNameIgnoreRegex.MatchString(repo)
		if !nameMatches {
			return true, nil
		}
		return false, nil
	}
}

// NewIPFilter is a default IPRemovalPredicate factory
// Image is matching the criteria if it's:
// - CreationTimestamp indicates that it is created more than ageInHours ago.
func NewImageFilter(ageInHours int) ImageRemovalPredicate {
	return func(image *gcrgoogle.ManifestInfo) bool {
		oldEnough := false

		imageAgeHours := time.Since(image.Created).Hours() - float64(ageInHours)
		oldEnough = imageAgeHours > 0

		if oldEnough {
			return true
		}
		return false
	}
}

/*
func (c *Cleaner) Clean(repo string, since time.Time, allowTagged bool, keep int, tagFilterRegexp *regexp.Regexp, dryRun bool) ([]string, error) {
	gcrrepo, err := gcrname.NewRepository(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo %s: %w", repo, err)
	}
	tags, err := gcrgoogle.List(gcrrepo, gcrgoogle.WithAuth(c.auther))
	if err != nil {
		return nil, fmt.Errorf("failed to list tags for repo %s: %w", repo, err)
	}

	// Create a worker pool for parallel deletion
	pool := workerpool.New(c.concurrency)

	var keepCount = 0
	var deleted = make([]string, 0, len(tags.Manifests))
	var deletedLock sync.Mutex
	var errs = make(map[string]error)
	var errsLock sync.RWMutex

	var manifests = make([]manifest, 0, len(tags.Manifests))
	for k, m := range tags.Manifests {
		manifests = append(manifests, manifest{k, m})
	}

	// Sort manifest by Created from the most recent to the least
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[j].Info.Created.Before(manifests[i].Info.Created)
	})

	for _, m := range manifests {
		if c.shouldDelete(m.Info, since, allowTagged, tagFilterRegexp) {
			// Keep a certain amount of images
			if keepCount < keep {
				keepCount++
				continue
			}

			// Deletes all tags before deleting the image
			for _, tag := range m.Info.Tags {
				tagged := gcrrepo.Tag(tag)
				if err := c.deleteOne(tagged, dryRun); err != nil {
					return nil, fmt.Errorf("failed to delete %s: %w", tagged, err)
				}
			}

			ref := gcrrepo.Digest(m.Digest)
			pool.Submit(func() {
				// Do not process if previous invocations failed. This prevents a large
				// build-up of failed requests and rate limit exceeding (e.g. bad auth).
				errsLock.RLock()
				if len(errs) > 0 {
					errsLock.RUnlock()
					return
				}
				errsLock.RUnlock()

				if err := c.deleteOne(ref, dryRun); err != nil {
					cause := errors.Unwrap(err).Error()

					errsLock.Lock()
					if _, ok := errs[cause]; !ok {
						errs[cause] = err
						errsLock.Unlock()
						return
					}
					errsLock.Unlock()
				}

				deletedLock.Lock()
				deleted = append(deleted, m.Digest)
				deletedLock.Unlock()
			})
		}
	}

	// Wait for everything to finish
	pool.StopWait()

	// Aggregate any errors
	if len(errs) > 0 {
		var errStrings []string
		for _, v := range errs {
			errStrings = append(errStrings, v.Error())
		}

		if len(errStrings) == 1 {
			return nil, fmt.Errorf(errStrings[0])
		}

		return nil, fmt.Errorf("%d errors occurred: %s",
			len(errStrings), strings.Join(errStrings, ", "))
	}

	return deleted, nil
}
*/
