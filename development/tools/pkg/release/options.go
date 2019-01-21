package release

import (
	"bytes"
	"context"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/file"
	"github.com/pkg/errors"
)

//Options represents the query options to create a Github release
type Options struct {
	Version      string
	Body         string
	TargetCommit string
	IsPreRelease bool
	storage      StorageAPI
}

//NewOptions returns new instance of Options
func NewOptions(ctx context.Context, storage StorageAPI, releaseVersionFilePath, releaseChangelogName, commitish string) (*Options, error) {

	relOpts := &Options{
		storage:      storage,
		TargetCommit: commitish,
	}

	releaseVersion, isPreRelease, err := readReleaseVersion(releaseVersionFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading %s file", releaseVersionFilePath)
	}

	//Changelog
	releaseChangelogData, err := relOpts.readReleaseBody(ctx, releaseVersion, releaseChangelogName)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading %s file", releaseChangelogName)
	}

	relOpts.Version = releaseVersion
	relOpts.IsPreRelease = isPreRelease
	relOpts.Body = releaseChangelogData

	return relOpts, nil

}

func readReleaseVersion(releaseVersionFilePath string) (string, bool, error) {

	common.Shout("Reading release version file")

	releaseVersion, err := file.ReadFile(releaseVersionFilePath)
	if err != nil {
		return "", false, err
	}

	isPreRelease := strings.Contains(releaseVersion, "rc")

	common.Shout("Release version: %s, Pre-release: %t", releaseVersion, isPreRelease)

	return releaseVersion, isPreRelease, nil
}

func (ro *Options) readReleaseBody(ctx context.Context, releaseVersion, releaseChangelogName string) (string, error) {

	releaseChangelogFullName := releaseVersion + "/" + releaseChangelogName

	releaseChangelogData, _,  err := ro.storage.ReadBucketObject(ctx, releaseChangelogFullName)
	if err != nil {
		return "", nil
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(releaseChangelogData)

	return buf.String(), nil
}
