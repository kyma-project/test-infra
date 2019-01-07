package githubrelease

import (
	"io/ioutil"
	"os"
)

// Release .
type Release struct {
	Gap *GithubAPIWrapper
	Saw *StorageAPIWrapper
}

//CreateRelease .
func (gr *Release) CreateRelease(releaseVersion string, targetCommit string, releaseChangelogName string, localArtifactName string, clusterArtifactName string, isPreRelease bool) error {
	//kymaConfigCluster
	clusterArtifactFile, err := gr.Saw.ReadBucketObject(clusterArtifactName)
	if err != nil {
		return err
	}

	//kymaConfigLocal
	localArtifactFile, err := gr.Saw.ReadBucketObject(localArtifactName)
	if err != nil {
		return err
	}

	//changelog
	releaseChangelogFile, err := gr.Saw.ReadBucketObject(releaseChangelogName)
	if err != nil {
		return err
	}

	releaseChangelogString, err := readFile(releaseChangelogFile.Name())
	if err != nil {
		return err
	}

	//release
	release, _, err := gr.Gap.CreateGithubRelease(releaseVersion, releaseChangelogString, targetCommit, isPreRelease)
	if err != nil {
		return err
	}

	//localArtifactFile
	_, _, err = gr.Gap.UploadArtifact(*release.ID, localArtifactName, localArtifactFile)
	if err != nil {
		return err
	}

	//clusterArtifactFile
	_, _, err = gr.Gap.UploadArtifact(*release.ID, clusterArtifactName, clusterArtifactFile)
	if err != nil {
		return err
	}

	return nil
}

func readFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	content := string(data[:])

	return content, nil
}
