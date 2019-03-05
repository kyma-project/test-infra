package syncomponent

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Component represents the element for which version will be checked
type Component struct {
	Name       string
	Path       string
	GitHash    string
	CommitDate string
	outOfDate  bool
	Versions   []*ComponentVersions
}

// ComponentVersions is a part of Component which includes information about version
type ComponentVersions struct {
	VersionPath   string
	Version       string
	ModifiedFiles []string
	outOfDate     bool
}

// NewSynComponent returns new Component element
func NewSynComponent(componentPath string, versionPaths []string) *Component {
	componentVersions := []*ComponentVersions{}
	for _, versionPath := range versionPaths {
		componentVersion := &ComponentVersions{VersionPath: versionPath}
		componentVersions = append(componentVersions, componentVersion)
	}

	nameElements := strings.Split(componentPath, "/")
	nameDash := nameElements[len(nameElements)-1]
	name := strings.Replace(nameDash, "-", "_", -1)

	return &Component{
		Name:     name,
		Path:     componentPath,
		Versions: componentVersions,
	}
}

// CheckIsOutOfDate determines whether a given component is not expired
func (c *Component) CheckIsOutOfDate(outOfDateDays int) {
	days := daysDelta(c.CommitDate, time.Now().Unix())
	if days <= outOfDateDays {
		log.Printf("Component %q does not achive the limit of days %d (limit is %d)", c.Name, days, outOfDateDays)
		return
	}

	for _, ver := range c.Versions {
		ver.checkIsOutOfDate(c.GitHash)
		c.outOfDate = ver.outOfDate
	}
}

func (cv *ComponentVersions) checkIsOutOfDate(hash string) {
	cut := hash[:len(cv.Version)]
	if cut == cv.Version {
		return
	}

	foundSourceCodeFiles := false
	regexPatterrn := regexp.MustCompile(`^.*\.(md|png|svg|dia)$`)
	for _, file := range cv.ModifiedFiles {
		if !regexPatterrn.Match([]byte(file)) {
			foundSourceCodeFiles = true
			break
		}
	}
	// if files like .md or images were modified then skip
	if !foundSourceCodeFiles {
		return
	}

	cv.outOfDate = true
}

func daysDelta(unixString string, unixToday int64) int {
	unixHashCommit, err := strconv.ParseInt(unixString, 10, 64)
	if err != nil {
		log.Fatalf("Cannot convert hash commit date %q to unix time: %s", unixString, err)
	}

	return int((unixToday - unixHashCommit) / (60 * 60 * 24))
}
