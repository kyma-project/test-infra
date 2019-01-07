package githubrelease

import (
	"github.com/kyma-project/test-infra/development/tools/pkg/file"
	"github.com/pkg/errors"
)

// Release .
type Release struct {
	Github  *GithubAPIWrapper
	Storage *StorageAPIWrapper
}

//CreateRelease .
func (gr *Release) CreateRelease(releaseVersion string, targetCommit string, releaseChangelogName string, localArtifactName string, clusterArtifactName string, isPreRelease bool) error {
	//kymaConfigCluster
	clusterArtifactFile, err := gr.Storage.ReadBucketObject(clusterArtifactName)
	if err != nil {
		return errors.Wrapf(err, "while reading %s file", clusterArtifactName)
	}

	//kymaConfigLocal
	localArtifactFile, err := gr.Storage.ReadBucketObject(localArtifactName)
	if err != nil {
		return errors.Wrapf(err, "while reading %s file", localArtifactName)
	}

	//changelog
	releaseChangelogFile, err := gr.Storage.ReadBucketObject(releaseChangelogName)
	if err != nil {
		return errors.Wrapf(err, "while reading %s file", releaseChangelogName)
	}

	releaseChangelogString, err := file.ReadFile(releaseChangelogFile.Name())
	if err != nil {
		return errors.Wrapf(err, "while reading %s file", releaseChangelogName)
	}

	//release
	release, _, err := gr.Github.CreateGithubRelease(releaseVersion, releaseChangelogString, targetCommit, isPreRelease)
	if err != nil {
		return errors.Wrapf(err, "while creating github release")
	}

	//localArtifactFile
	_, _, err = gr.Github.UploadArtifact(*release.ID, localArtifactName, localArtifactFile)
	if err != nil {
		return errors.Wrapf(err, "while uploading %s artifact to %s release", localArtifactName, *release.ID)
	}

	//clusterArtifactFile
	_, _, err = gr.Github.UploadArtifact(*release.ID, clusterArtifactName, clusterArtifactFile)
	if err != nil {
		return errors.Wrapf(err, "while uploading %s artifact to %s release", clusterArtifactName, *release.ID)
	}

	return nil
}
