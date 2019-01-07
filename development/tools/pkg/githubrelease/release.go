package githubrelease

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/pkg/errors"
)

// Release .
type Release struct {
	Github  *GithubAPIWrapper
	Storage *StorageAPIWrapper
}

var kymaArtifactsDir = "kyma-artifacts"

//CreateRelease .
func (gr *Release) CreateRelease(releaseVersion string, targetCommit string, releaseChangelogName string, localArtifactName string, clusterArtifactName string, isPreRelease bool) error {
	artifactsDir, err := ioutil.TempDir("", kymaArtifactsDir)
	if err != nil {
		log.Fatal(err)
	}

	//kymaConfigCluster
	clusterArtifactData, err := gr.Storage.ReadBucketObject(clusterArtifactName)
	if err != nil {
		return errors.Wrapf(err, "while reading %s from bucket", clusterArtifactName)
	}

	clusterArtifactFile, err := os.OpenFile(path.Join(artifactsDir, clusterArtifactName), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.Wrapf(err, "while opening %s file", clusterArtifactName)
	}

	clusterArtifactFile, err = saveArtifact(clusterArtifactFile, clusterArtifactData)
	if err != nil {
		return errors.Wrapf(err, "while writing %s file", clusterArtifactName)
	}

	//kymaConfigLocal
	localArtifactData, err := gr.Storage.ReadBucketObject(localArtifactName)
	if err != nil {
		return errors.Wrapf(err, "while reading %s file", localArtifactName)
	}

	localArtifactFile, err := os.OpenFile(path.Join(artifactsDir, localArtifactName), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.Wrapf(err, "while opening %s file", clusterArtifactName)
	}

	localArtifactFile, err = saveArtifact(localArtifactFile, localArtifactData)
	if err != nil {
		return errors.Wrapf(err, "while writing %s file", localArtifactFile)
	}

	defer func() {
		clusterArtifactFile.Close()
		localArtifactFile.Close()
		os.RemoveAll(artifactsDir)
	}()

	//changelog
	releaseChangelogData, err := gr.Storage.ReadBucketObject(releaseChangelogName)
	if err != nil {
		return errors.Wrapf(err, "while reading %s file", releaseChangelogName)
	}

	//release
	release, _, err := gr.Github.CreateGithubRelease(releaseVersion, string(releaseChangelogData), targetCommit, isPreRelease)
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

func saveArtifact(artifactFile *os.File, artifactData []byte) (*os.File, error) {
	_, err := artifactFile.Write(artifactData)
	if err != nil {
		return nil, err
	}

	err = artifactFile.Close()
	if err != nil {
		return nil, err
	}

	artifactFile, err = os.Open(artifactFile.Name())
	if err != nil {
		return nil, err
	}

	return artifactFile, nil
}
