package release

import (
	"regexp"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/file"
	"github.com/pkg/errors"
)

// VersionReader wraps the Read method that reads RELEASE_VERSION file
type VersionReader interface {
	Read(filePath string) (string, bool, error)
}

type kymaVersionReader struct{}

// NewVersionReader returns a ready-to-use implementation of VersionReadeer
func NewVersionReader() VersionReader {
	return &kymaVersionReader{}
}

func (*kymaVersionReader) Read(filePath string) (string, bool, error) {

	common.Shout("Reading release version file")

	releaseVersion, err := file.ReadFile(filePath)
	if err != nil {
		return "", false, err
	}

	r, err := regexp.Compile("^(\\d.){2}\\d(-rc)?$")
	if err != nil {
		return "", false, err
	}

	if !r.MatchString(releaseVersion) {
		return "", false, errors.New("Kyma version provided in the RELASE_VERSION file is malformed")
	}

	isPreRelease := strings.Contains(releaseVersion, "rc")

	common.Shout("Release version: %s, Pre-release: %t", releaseVersion, isPreRelease)

	return releaseVersion, isPreRelease, nil

}
