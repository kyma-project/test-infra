package release

import (
	"bytes"
	"context"
	"os"

	"github.com/pkg/errors"
)

// Options represents the query options to create a Github release
type Options struct {
	Version      string
	Body         string
	TargetCommit string
	IsPreRelease bool
	storage      StorageAPI
}

// NewOptions returns new instance of Options
func NewOptions(ctx context.Context, storage StorageAPI, releaseVersionFilePath, releaseChangelogName, commitish string, r VersionReader) (*Options, error) {

	relOpts := &Options{
		storage:      storage,
		TargetCommit: commitish,
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
	if _, err := os.Stat(releaseChangelogName); err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "while reading %s file", releaseChangelogName)
		}
		// use try to use GCP to fetch the changelog
		releaseChangelogData, err = relOpts.readReleaseBody(ctx, releaseVersion, releaseChangelogName)
		if err != nil {
			return nil, errors.Wrapf(err, "while reading %s file", releaseChangelogName)
		}
	} else {
		clb, err := os.ReadFile(releaseChangelogName)
		if err != nil {
			return nil, errors.Wrapf(err, "while reading %s file", releaseChangelogName)
		}
		releaseChangelogData = string(clb)
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

	defer releaseChangelogData.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(releaseChangelogData)

	return buf.String(), nil
}
