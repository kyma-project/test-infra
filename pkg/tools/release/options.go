package release

import (
	"os"

	"github.com/pkg/errors"
)

// Options represents the query options to create a Github release
type Options struct {
	Version            string
	Body               string
	TargetCommit       string
	IsPreRelease       bool
	KymaComponentsName string
	KymaComponentsPath string
}

// NewOptions returns new instance of Options
func NewOptions(releaseVersionFilePath, releaseChangelogName, commitSHA, kymaComponentsName, kymaComponentsPath string, r VersionReader) (*Options, error) {

	relOpts := &Options{
		TargetCommit:       commitSHA,
		KymaComponentsName: kymaComponentsName,
		KymaComponentsPath: kymaComponentsPath,
	}

	if r == nil {
		r = NewVersionReader()
	}
	releaseVersion, isPreRelease, err := r.ReadFromFile(releaseVersionFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading %s file", releaseVersionFilePath)
	}

	//Changelog
	var releaseChangelogData string
	_, err = os.Stat(releaseChangelogName)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading %s file", releaseChangelogName)
	}
	clb, err := os.ReadFile(releaseChangelogName)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading %s file", releaseChangelogName)
	}
	releaseChangelogData = string(clb)

	relOpts.Version = releaseVersion
	relOpts.IsPreRelease = isPreRelease
	relOpts.Body = releaseChangelogData

	return relOpts, nil

}
