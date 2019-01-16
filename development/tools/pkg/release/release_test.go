package release_test

import (
	"context"
	"testing"

	. "github.com/kyma-project/test-infra/development/tools/pkg/release"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateRelease(t *testing.T) {

	fakeGithub := &FakeGithubAPIWrapper{}
	fakeStorage := &FakeStorageAPIWrapper{}

	releaseWizard := NewCreator(fakeGithub, fakeStorage)

	Convey("If provided with correct data", t, func() {

		//given
		ctx := context.Background()
		mockRelVer := "0.0.1"
		mockCommitish := "a1b2c3d4"
		mockChangelogFileName := "change-record.md"
		expectedBody := "test artifact data for change-record.md"
		mockLocalArtifactName := "kyma-config-local.yaml"
		mockClusterArtifactName := "kyma-config-cluster.yaml"
		mockIsPreRelease := false

		Convey("CreateNewRelease function should download three files from Google Storage, create a release and upload two assets", func() {

			//when
			err := releaseWizard.CreateNewRelease(ctx, mockRelVer, mockCommitish, mockChangelogFileName, mockLocalArtifactName, mockClusterArtifactName, mockIsPreRelease)

			//then
			So(fakeStorage.TimesReadBucketObjectCalled, ShouldEqual, 3)

			So(err, ShouldBeNil)
			So(fakeGithub.Release.GetBody(), ShouldEqual, expectedBody)

			So(fakeGithub.TimesUploadFileCalled, ShouldEqual, 2)
			So(fakeGithub.AssetCount, ShouldEqual, 2)
			So(fakeGithub.Assets[0].GetName(), ShouldEqual, mockLocalArtifactName)
			So(fakeGithub.Assets[1].GetName(), ShouldEqual, mockClusterArtifactName)

		})
	})
}
