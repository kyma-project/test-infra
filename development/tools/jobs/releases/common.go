package releases

import (
	"fmt"
	"github.com/Masterminds/semver"
)

// SupportedRelease defines supported releases
type SupportedRelease semver.Version

// GetKymaReleasesUntil filters all available releases earlier or the same as the given one
func GetKymaReleasesUntil(lastRelease *SupportedRelease) []*SupportedRelease {
	var supportedReleases []*SupportedRelease

	for _, rel := range GetAllKymaReleases() {
		if rel.Compare(lastRelease) <= 0 {
			supportedReleases = append(supportedReleases, rel)
		}
	}

	return supportedReleases
}

// GetKymaReleasesSince filters all available releases later or the same as the given one
func GetKymaReleasesSince(firstRelease *SupportedRelease) []*SupportedRelease {
	var supportedReleases []*SupportedRelease

	for _, rel := range GetAllKymaReleases() {
		if rel.Compare(firstRelease) >= 0 {
			supportedReleases = append(supportedReleases, rel)
		}
	}

	return supportedReleases
}

// Compare compares this version to another one. It returns -1, 0, or 1 if
// the version smaller, equal, or larger than the other version.
func (r *SupportedRelease) Compare(other *SupportedRelease) int {
	return (*semver.Version)(r).Compare((*semver.Version)(other))
}

// Branch returns a git branch for this release
func (r *SupportedRelease) Branch() string {
	return fmt.Sprintf("release-%v.%v", (*semver.Version)(r).Major(), (*semver.Version)(r).Minor())
}

// JobPrefix returns a prefix for all jobs for this release
func (r *SupportedRelease) JobPrefix() string {
	return fmt.Sprintf("rel%v%v", (*semver.Version)(r).Major(), (*semver.Version)(r).Minor())
}

// String returns formatted release
func (r *SupportedRelease) String() string {
	return fmt.Sprintf("%v.%v", (*semver.Version)(r).Major(), (*semver.Version)(r).Minor())
}

func mustParse(v string) *SupportedRelease {
	parsed := SupportedRelease(*semver.MustParse(v))
	return &parsed
}
