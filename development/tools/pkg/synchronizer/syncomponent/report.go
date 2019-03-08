package syncomponent

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
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
		log.Println(outOfDateVersionLog(c))

		reports = append(reports, Report{
			componentName: prettyComponentName(c.Name),
			message:       prettyMessage(c),
		})
	}
	log.Printf("There are %d components with alerts \n", len(reports))

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

type componentLog struct {
	elements map[int]string
}

func newComponentLog() *componentLog {
	return &componentLog{
		elements: make(map[int]string),
	}
}

func (cl *componentLog) addElement(order int, element string) {
	cl.elements[order] = element
}

func (cl componentLog) log() string {
	var keys []int
	for k := range cl.elements {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	response := ""
	for _, k := range keys {
		response += cl.elements[k] + "\n"
	}

	return response
}

func (cl *componentLog) addCurrentHashMessage(order int, c *Component) {
	var versionMsg []string
	for _, ver := range c.Versions {
		versionMsg = append(versionMsg, fmt.Sprintf("versions: %s", ver.Version))
	}
	versionMessage := strings.Join(versionMsg, ",")

	cl.addElement(order, fmt.Sprintf("Component hash: %s, component %s", c.GitHash, versionMessage))
}

func currentVersionLog(c *Component) string {
	cLog := newComponentLog()
	cLog.addElement(1, fmt.Sprintf("Component %q is not expired.", c.Name))
	cLog.addCurrentHashMessage(2, c)

	gitHistoryLog := prettyVersionContainsGitHistory(c)
	// if information about contains in git history log is not empty print it
	// if is empty print information about commited files
	if gitHistoryLog != "" {
		cLog.addElement(3, gitHistoryLog)
	} else {
		cLog.addElement(4, prettyFilesExstensionList(c.Versions))
	}

	return cLog.log()
}

func outOfDateVersionLog(c *Component) string {
	cLog := newComponentLog()
	cLog.addElement(1, fmt.Sprintf("Component %q is out of date.", c.Name))
	cLog.addCurrentHashMessage(2, c)
	cLog.addElement(3, prettyFilesExstensionList(c.Versions))

	return cLog.log()
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

func prettyFilesExstensionList(versions []*ComponentVersions) string {
	data := make(map[string][]string, len(versions))

	ext := func(files []string) []string {
		keys := map[string]bool{}
		resp := []string{}
		for _, file := range files {
			ext := filepath.Ext(file)
			if ext == "" {
				ext = filepath.Base(file)
			}
			if _, value := keys[ext]; value {
				continue
			}
			keys[ext] = true
			if ext == "." {
				continue
			}
			resp = append(resp, ext)
		}
		return resp
	}

	for _, version := range versions {
		data[version.VersionPath] = ext(version.ModifiedFiles)
	}

	var response string
	printMsg := false
	for name, files := range data {
		extensions := strings.Join(files, ", ")
		if extensions != "" {
			printMsg = true
		}
		response += fmt.Sprintf("\n for version in resource %q: [%s]", name, extensions)
	}

	if !printMsg {
		return ""
	}
	return fmt.Sprintf("File extensions that have been changed since last allowed commit: %s", response)
}

func prettyTime(unix int64) string {
	tm := time.Unix(unix, 0)

	return fmt.Sprintf("%d %s %d %d:%d:%d", tm.Day(), tm.Month(), tm.Year(), tm.Hour(), tm.Minute(), tm.Second())
}
