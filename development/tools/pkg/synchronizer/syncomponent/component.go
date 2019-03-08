package syncomponent

import (
	"math"
	"regexp"
	"strings"
)

// Component represents the element for which version will be checked
type Component struct {
	Name           string
	Path           string
	GitHash        string
	GitHashHistory map[int64]string
	outOfDate      bool
	Versions       []*ComponentVersions
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

	history := make(map[int64]string)

	return &Component{
		Name:           name,
		Path:           componentPath,
		GitHashHistory: history,
		Versions:       componentVersions,
	}
}

// GetOldestAllowed returns the oldest allowed hash commit of component
func (c Component) GetOldestAllowed() string {
	if len(c.GitHashHistory) == 0 {
		return c.GitHash
	}

	oldest := int64(math.MaxInt64)
	for key := range c.GitHashHistory {
		if oldest > key {
			oldest = key
		}
	}

	return c.GitHashHistory[oldest]
}

// CheckIsOutOfDate determines whether a given component is not expired
func (c *Component) CheckIsOutOfDate() {
	for _, ver := range c.Versions {
		ver.checkIsOutOfDate(c.GitHash, c.GitHashHistory)
		c.outOfDate = ver.outOfDate
	}
}

func (cv *ComponentVersions) checkIsOutOfDate(currentHash string, hashHistory map[int64]string) {
	cut := currentHash[:len(cv.Version)]
	if cut == cv.Version {
		return
	}

	for _, hash := range hashHistory {
		cut = hash[:len(cv.Version)]
		if cv.Version == cut {
			return
		}
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
