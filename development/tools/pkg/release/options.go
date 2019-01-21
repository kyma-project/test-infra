package release

import (
	"bytes"
	"context"

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
func NewOptions(ctx context.Context, storage StorageAPI, releaseVersionFilePath, releaseChangelogName, commitish string, r VersionReader) (*Options, error) {

	relOpts := &Options{
		storage:      storage,
		TargetCommit: commitish,
	}

	if r == nil {
		r = NewVersionReader()
	}

	releaseVersion, isPreRelease, err := r.Read(releaseVersionFilePath)
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

func (ro *Options) readReleaseBody(ctx context.Context, releaseVersion, releaseChangelogName string) (string, error) {

	releaseChangelogFullName := releaseVersion + "/" + releaseChangelogName

	releaseChangelogData, _, err := ro.storage.ReadBucketObject(ctx, releaseChangelogFullName)
	if err != nil {
		return "", nil
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(releaseChangelogData)

	return buf.String(), nil
}
