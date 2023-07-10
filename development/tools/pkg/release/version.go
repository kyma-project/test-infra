package release

import (
	"regexp"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/file"
	"github.com/pkg/errors"
)

// VersionReader wraps the ReadFromFile method that reads RELEASE_VERSION file
type VersionReader interface {
	ReadFromFile(filePath string) (string, bool, error)
}

type kymaVersionReader struct{}

// NewVersionReader returns a ready-to-use implementation of VersionReadeer
func NewVersionReader() VersionReader {
	return &kymaVersionReader{}
}

// ReadFromFile reads the file that contains Kyma version and returns it. It returns true if it is a pre-release version.
func (*kymaVersionReader) ReadFromFile(filePath string) (string, bool, error) {

	common.Shout("Reading release version file")

	releaseVersion, err := file.ReadFile(filePath)
	if err != nil {
		return "", false, err
	}

	isPreRelease, err := parseVersion(releaseVersion)
	if err != nil {
		return "", false, err
	}

	common.Shout("Release version: %s, Pre-release: %t", releaseVersion, isPreRelease)

	return releaseVersion, isPreRelease, nil

}

func parseVersion(releaseVersion string) (bool, error) {

	r, err := regexp.Compile(`^(\d+.){2}\d+(-rc(\d+)?)?$`)
	if err != nil {
		return false, err
	}

	if !r.MatchString(strings.TrimSpace(releaseVersion)) {
		return false, errors.New("Kyma version provided in the RELEASE_VERSION file is malformed")
	}

	return strings.Contains(releaseVersion, "rc"), nil
}
