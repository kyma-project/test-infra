package syncomponent

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaultOutOfDateDays = 3

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

// SetOutOfDate determines whether a given component is not expired
func (c *Component) SetOutOfDate(outOfDateDays int) {
	days := daysDelta(c.CommitDate, time.Now().Unix())
	if outOfDateDays == 0 {
		outOfDateDays = defaultOutOfDateDays
	}
	if days <= outOfDateDays {
		log.Printf("Component %q does not achive the limit of days %d (limit is %d)", c.Name, days, outOfDateDays)
		return
	}

	for _, ver := range c.Versions {
		ver.setOutOfDate(c.GitHash)
		c.outOfDate = ver.outOfDate
	}
}

func (cv *ComponentVersions) setOutOfDate(hash string) {
	cut := hash[:len(cv.Version)]
	if cut == cv.Version {
		return
	}

	followedFiles := false
	regexPatterrn := regexp.MustCompile(`^.*\.(md|png|svg|dia)$`)
	for _, file := range cv.ModifiedFiles {
		if !regexPatterrn.Match([]byte(file)) {
			followedFiles = true
			break
		}
	}
	// if files like .md or images were modified then skip
	if !followedFiles {
		return
	}

	cv.outOfDate = true
}

// RelativePathToComponent returns path to directory exclude root path
func RelativePathToComponent(mainPath string, componenthPath string) string {
	re := regexp.MustCompile(mainPath)
	componentPath := re.ReplaceAllLiteralString(componenthPath, "")

	return strings.Trim(componentPath, "/")
}

func daysDelta(unixString string, unixToday int64) int {
	unixHashCommit, err := strconv.ParseInt(unixString, 10, 64)
	if err != nil {
		log.Fatalf("Cannot convert hash commit date %q to unix time: %s", unixString, err)
	}

	return int((unixToday - unixHashCommit) / (60 * 60 * 24))
}
