package release_test

import (
	"context"
	"testing"

	. "github.com/kyma-project/test-infra/development/tools/pkg/release"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	mockLocalArtifactName   = "mockcluster.yaml"
	mockClusterArtifactName = "kyma-config-cluster.yaml"
	mockChangelogFileName   = "change-record.md"
	mockCommitish           = "a1b2c3d4"
)

func TestCreateRelease(t *testing.T) {

	Convey("CreateNewRelease func", t, func() {

		ctx := context.Background()
		vReader := &FakeKymaVersionReader{}

		Convey("provided with correct release data", func() {

			//given
			fakeGithub := &FakeGithubAPIWrapper{}
			fakeStorage := &FakeStorageAPIWrapper{}
			releaseWizard := NewCreator(fakeGithub, fakeStorage)

			mockRelVer := "0.0.1"
			expectedBody := "test artifact data for 0.0.1/change-record.md"

			relOpts, _ := NewOptions(ctx, fakeStorage, mockRelVer, mockChangelogFileName, mockCommitish, vReader)

			Convey("should download three files from Google Storage, create a release and upload two assets", func() {

				//when
				err := releaseWizard.CreateNewRelease(ctx, relOpts, mockLocalArtifactName, mockClusterArtifactName)

				//then
				So(fakeStorage.TimesReadBucketObjectCalled, ShouldEqual, 3)

				So(err, ShouldBeNil)
				So(fakeGithub.Release.GetBody(), ShouldEqual, expectedBody)
				So(fakeGithub.Release.GetPrerelease(), ShouldBeFalse)

				So(fakeGithub.TimesUploadFileCalled, ShouldEqual, 2)
				So(fakeGithub.AssetCount, ShouldEqual, 2)
				So(fakeGithub.Assets[0].GetName(), ShouldEqual, mockLocalArtifactName)
				So(fakeGithub.Assets[1].GetName(), ShouldEqual, mockClusterArtifactName)

			})
		})

		Convey("provided with correct pre-release data", func() {

			//given
			fakeGithub := &FakeGithubAPIWrapper{}
			fakeStorage := &FakeStorageAPIWrapper{}
			releaseWizard := NewCreator(fakeGithub, fakeStorage)

			mockRelVer := "0.0.2-rc"
			expectedBody := "test artifact data for 0.0.2-rc/change-record.md"

			relOpts, _ := NewOptions(ctx, fakeStorage, mockRelVer, mockChangelogFileName, mockCommitish, vReader)

			Convey("should download three files from Google Storage, create a pre-release and upload two assets", func() {

				//when
				err := releaseWizard.CreateNewRelease(ctx, relOpts, mockLocalArtifactName, mockClusterArtifactName)

				//then
				So(fakeStorage.TimesReadBucketObjectCalled, ShouldEqual, 3)

				So(err, ShouldBeNil)
				So(fakeGithub.Release.GetBody(), ShouldEqual, expectedBody)
				So(fakeGithub.Release.GetPrerelease(), ShouldBeTrue)

				So(fakeGithub.TimesUploadFileCalled, ShouldEqual, 2)
				So(fakeGithub.AssetCount, ShouldEqual, 2)
				So(fakeGithub.Assets[0].GetName(), ShouldEqual, mockLocalArtifactName)
				So(fakeGithub.Assets[1].GetName(), ShouldEqual, mockClusterArtifactName)

			})
		})
	})
}
