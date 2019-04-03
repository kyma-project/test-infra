package release

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseVersion(t *testing.T) {

	Convey("TestParseVersion func", t, func() {

		Convey("should return false and no error if the provided version is valid and refers to a full-release", func() {

			//given
			version := "0.0.4"

			//when
			isPreRel, err := parseVersion(version)

			//then
			So(err, ShouldBeNil)
			So(isPreRel, ShouldBeFalse)

		})

		Convey("should accept multi-digit numbers", func() {

			//given
			version := "10.204.35000"

			//when
			isPreRel, err := parseVersion(version)

			//then
			So(err, ShouldBeNil)
			So(isPreRel, ShouldBeFalse)

		})

		Convey("should trim newlines", func() {

			//given
			version := `10.204.35000
`

			//when
			isPreRel, err := parseVersion(version)

			//then
			So(err, ShouldBeNil)
			So(isPreRel, ShouldBeFalse)

		})

		Convey("should return true and no error if the provided version is valid and refers to a pre-release", func() {

			//given
			version := "0.0.3-rc"

			//when
			isPreRel, err := parseVersion(version)

			//then
			So(err, ShouldBeNil)
			So(isPreRel, ShouldBeTrue)

		})

		Convey("should return true and no error if case of multiple versions of the same pre-release", func() {

			//given
			version := "0.0.3-rc8"

			//when
			isPreRel, err := parseVersion(version)

			//then
			So(err, ShouldBeNil)
			So(isPreRel, ShouldBeTrue)

		})

		Convey("should return an error if the provided version is not in line with SemVer specification", func() {

			//given
			version := "0.2"

			//when
			_, err := parseVersion(version)

			//then
			So(err, ShouldNotBeNil)

		})

		Convey("should return an error if the provided version is malformed", func() {

			//given
			version := "0.0.1-zz"

			//when
			_, err := parseVersion(version)

			//then
			So(err, ShouldNotBeNil)

		})
	})
}
