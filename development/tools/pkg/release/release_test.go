package release_test

import (
	"context"
	"testing"

	. "github.com/kyma-project/test-infra/development/tools/pkg/release"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	mockChangelogFileName = "change-record.md"
	mockCommitiSHA        = "a1b2c3d4"
	mockComponentsPath    = "kyma-components.yaml"
)

func TestCreateRelease(t *testing.T) {

	Convey("CreateNewRelease func", t, func() {

		ctx := context.Background()
		vReader := &FakeKymaVersionReader{}

		Convey("provided with correct release data", func() {

			//given
			fakeGithub := &FakeGithubAPIWrapper{}
			releaseWizard := NewCreator(fakeGithub)

			mockRelVer := "0.0.1"

			relOpts, _ := NewOptions(mockRelVer, mockChangelogFileName, mockCommitiSHA, mockComponentsPath, mockComponentsPath, vReader)

			Convey("should create a release and upload two assets", func() {

				//when
				err := releaseWizard.CreateNewRelease(ctx, relOpts)

				//then
				So(err, ShouldBeNil)
				So(len(fakeGithub.Release.GetBody()), ShouldEqual, 95)
				So(fakeGithub.Release.GetPrerelease(), ShouldBeFalse)

				So(fakeGithub.TimesUploadFileCalled, ShouldEqual, 1)
				So(fakeGithub.AssetCount, ShouldEqual, 1)
				So(fakeGithub.Assets[0].GetName(), ShouldEqual, mockComponentsPath)
			})
		})

		Convey("provided with correct pre-release data", func() {

			//given
			fakeGithub := &FakeGithubAPIWrapper{}
			releaseWizard := NewCreator(fakeGithub)

			mockRelVer := "0.0.2-rc"

			relOpts, _ := NewOptions(mockRelVer, mockChangelogFileName, mockCommitiSHA, mockComponentsPath, mockComponentsPath, vReader)

			Convey("should create a pre-release and upload two assets", func() {

				//when
				err := releaseWizard.CreateNewRelease(ctx, relOpts)

				//then
				So(err, ShouldBeNil)
				So(len(fakeGithub.Release.GetBody()), ShouldEqual, 95)
				So(fakeGithub.Release.GetPrerelease(), ShouldBeTrue)

				So(fakeGithub.TimesUploadFileCalled, ShouldEqual, 1)
				So(fakeGithub.AssetCount, ShouldEqual, 1)
				So(fakeGithub.Assets[0].GetName(), ShouldEqual, mockComponentsPath)

			})
		})
	})
}
