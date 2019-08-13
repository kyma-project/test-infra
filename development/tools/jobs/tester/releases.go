package tester

import (
	"github.com/Masterminds/semver"
)

// List of currently supported releases
// Please always make it up to date
// When we removing support for given version, there remove
// its entry also here.
var (
	Release12 = mustParse("1.2")
	Release13 = mustParse("1.3")
	Release14 = mustParse("1.4")
)

// GetAllKymaReleaseBranches returns all supported kyma release branches
func GetAllKymaReleaseBranches() []*SupportedRelease {
	return []*SupportedRelease{Release12, Release13, Release14}
}


