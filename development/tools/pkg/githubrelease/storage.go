package githubrelease

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	"cloud.google.com/go/storage"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
)

// StorageAPIWrapper abstracts Google Storage API
type StorageAPIWrapper struct {
	Context       context.Context
	StorageClient *storage.Client
	BucketName    string
	FolderName    string
	TmpDir        string
}

// ReadBucketObject downloads and saves in temporary directory a file specified by fileName from a given bucket
func (saw *StorageAPIWrapper) ReadBucketObject(fileName string) (*os.File, error) {
	common.Shout("Reading %s/%s file from % bucket", saw.FolderName, fileName, saw.BucketName)

	rc, err := saw.StorageClient.Bucket(saw.BucketName).Object(saw.FolderName + "/" + fileName).NewReader(saw.Context)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return saveArtifact(saw.TmpDir, fileName, data)
}

func saveArtifact(artifactsDir string, artifactFileName string, artifactData []byte) (*os.File, error) {
	tmpArtifactFile, err := os.OpenFile(path.Join(artifactsDir, artifactFileName), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}

	_, err = tmpArtifactFile.Write(artifactData)
	if err != nil {
		return nil, err
	}

	err = tmpArtifactFile.Close()
	if err != nil {
		return nil, err
	}

	tmpArtifactFile, err = os.Open(tmpArtifactFile.Name())
	if err != nil {
		return nil, err
	}

	return tmpArtifactFile, nil
}
