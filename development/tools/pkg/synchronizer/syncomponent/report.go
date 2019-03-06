package syncomponent

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
)

// Report represents message about out of dates components
type Report struct {
	componentName string
	message       string
}

// GetTitle returns title report - component name
func (r Report) GetTitle() string {
	return r.componentName
}

// GetValue returns information about out of date component
func (r Report) GetValue() string {
	return r.message
}

// GenerateReport generates message about out of date components
func GenerateReport(components []*Component) []Report {
	var reports []Report

	log.Printf("There are %d components \n", len(components))
	for _, c := range components {
		if !c.outOfDate {
			log.Println(currentVersionLog(c))
			continue
		}

		reports = append(reports, Report{
			componentName: prettyComponentName(c.Name),
			message:       prettyMessage(c),
		})
	}

	return reports
}

func prettyComponentName(name string) string {
	return strings.Replace(name, "_", " ", -1)
}

func prettyMessage(c *Component) string {
	var parts []string

	for _, version := range c.Versions {
		parts = append(parts, fmt.Sprintf(
			"The version of the _%q_ component is *%s*", version.VersionPath, version.Version))
	}
	parts = append(parts, fmt.Sprintf(
		"The current component commit is *%s*",
		c.GitHash[:8],
	))

	return strings.Join(parts, "\n")
}

func currentVersionLog(c *Component) string {
	var versionMsg []string
	for _, ver := range c.Versions {
		versionMsg = append(versionMsg, fmt.Sprintf("versions: %s", ver.Version))
	}

	return fmt.Sprintf(
		"Component %q is not expired. \n"+
			"Component hash: %s, component %s"+
			prettyVersionContainsGitHistory(c)+
			prettyFilesExstensionList(c.Versions),
		c.Name,
		c.GitHash,
		strings.Join(versionMsg, ","),
	)
}

func prettyVersionContainsGitHistory(c *Component) string {
	parts := []string{}

	for _, ver := range c.Versions {
		cut := c.GitHash[:len(ver.Version)]
		if cut == ver.Version {
			continue
		}
		for date, hash := range c.GitHashHistory {
			cut = hash[:len(ver.Version)]
			if ver.Version == cut {
				parts = append(parts, fmt.Sprintf(
					"The version %q is included in the permitted log history: %s from %s",
					ver.Version,
					cut,
					prettyTime(date),
				))
			}
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("(%s)", strings.Join(parts, ","))
}

func prettyTime(unix int64) string {
	tm := time.Unix(unix, 0)

	return fmt.Sprintf("%d %s %d %d:%d:%d", tm.Day(), tm.Month(), tm.Year(), tm.Hour(), tm.Minute(), tm.Second())
}

func prettyFilesExstensionList(versions []*ComponentVersions) string {
	data := make(map[string][]string, len(versions))

	ext := func(files []string) []string {
		resp := []string{}
		for _, file := range files {
			resp = append(resp, filepath.Ext(file))
		}
		return resp
	}

	for _, version := range versions {
		data[version.VersionPath] = ext(version.ModifiedFiles)
	}

	var response string
	printMsg := false
	for name, files := range data {
		extensions := strings.Join(files, ",")
		if extensions != "" {
			printMsg = true
		}
		response += fmt.Sprintf("for version in resource %q: [%s]", name, extensions)
	}

	if !printMsg {
		return ""
	}
	return fmt.Sprintf("\n File extensions that have been changed: %s \n", response)
}
