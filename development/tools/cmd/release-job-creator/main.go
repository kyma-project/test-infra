package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kelseyhightower/envconfig"
	// "github.com/ghodss/yaml"
)

// func gatherOptions() options {
// 	o := options{}
// 	flag.StringVar(&o.jobType, "job-type", "all", "Type of job")
// 	flag.StringVar(&o.unsupportedReleases, "unsupported-releases", "", "Unsupported releases.")
// 	flag.StringVar(&o.newRelease, "new-release", "", "Name of the new release.")
// 	flag.Parse()
// 	return o
// }

const (
	jobsDir       = "prow/jobs"
	testInfraDir  = "src/github.com/kyma-project/test-infra"
	yamlExtension = ".yaml"
)

type envConfig struct {
	// GOPATH
	GoPath      string `envconfig:"GOPATH" default:"" required:"true"`
	OldReleases string `envconfig:"OLD_RELEASES" default:"" required:"true"`
	NewReleases string `envconfig:"NEW_RELEASES" default:"" required:"true"`
	RefRelease  string `envconfig:"REF_RELEASE" default:"" required:"true"`
}

var (
	components = []string{
		"kyma", "incubator/compass",
	}
	env envConfig
)

func main() {
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}
	os.Exit(_main(os.Args[1:], env))

}

func _main(args []string, env envConfig) int {
	for _, component := range components {
		processJobDef(env.GoPath + "/" + testInfraDir + "/" + jobsDir + "/" + component)
	}
	return 0
}

func createJobName(typ, rel, folder string) string {
	rel = strings.Replace(rel, ".", "", -1)

	return typ + "-" + "rel" + rel + "-" + folder
}

func processJobDef(root string) {
	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(info.Name(), yamlExtension) {
				fmt.Println("Processed file: " + path)
				fp := FileProcessor{
					fileName: path,
				}
				err := fp.readFileWithReadLine()
				if err != nil {
					log.Panicf(err.Error())
				}
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}

// FileProcessor holds details needed to process a job definition file
type FileProcessor struct {
	fileName          string
	finalContent      string
	oldContent        string
	sampleReleaseDefs []string
}

func (fp FileProcessor) readFileWithReadLine() (err error) {
	file, err := os.Open(fp.fileName)
	defer file.Close()
	addToFinalContent := true
	refIndicator := false
	sampleReleaseDef := ""
	if err != nil {
		return err
	}
	fp.finalContent = ""
	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)
	oldReleaseDef := false
	numOfExtractsForOldRelease := 0
	leadingSpaces := 0
	for {
		var buffer bytes.Buffer

		var l []byte
		var isPrefix bool
		for {
			l, isPrefix, err = reader.ReadLine()
			buffer.Write(l)
			// End of the line, stop reading.
			if !isPrefix {
				break
			}
			// At the EOF, break
			if err != nil {
				break
			}
		}
		if err == io.EOF {
			break
		}
		line := buffer.String()
		fp.oldContent += line + "\n"
		// The marks the end of a job definition
		if countLeadingSpaces(line) <= leadingSpaces {
			oldReleaseDef = false
			addToFinalContent = true
			if refIndicator {
				fp.sampleReleaseDefs = append(fp.sampleReleaseDefs, sampleReleaseDef)
				sampleReleaseDef = ""
			}
			refIndicator = false
		}

		if strings.Contains(line, "- name: pre-rel"+strings.ReplaceAll(env.RefRelease, ".", "")) {
			refIndicator = true
		}

		if refIndicator {
			sampleReleaseDef += line + "\n"
		}

		// Process the line here.
		for _, rel := range strings.Split(env.OldReleases, ",") {
			relWithoutDot := strings.ReplaceAll(rel, ".", "")
			if strings.Contains(line, "- name: pre-rel"+relWithoutDot) {
				leadingSpaces = countLeadingSpaces(line)
				addToFinalContent = false
				oldReleaseDef = true
				numOfExtractsForOldRelease++
				// if numOfExtractsForOldRelease == 1 {
				// 	fp.sampleReleaseNum = rel
				// }
				break
			}
			if oldReleaseDef {
				break
			}
		}

		if addToFinalContent {
			fp.finalContent += line + "\n"
		}

	}

	fp.finalContent = fp.addNewRelease()
	fp.overwriteFile()

	if err != io.EOF {
		fmt.Printf(" > Failed!: %v\n", err)
	}

	return nil
}

func (fp FileProcessor) overwriteFile() {
	d1 := []byte(fp.finalContent)
	err := ioutil.WriteFile(fp.fileName, d1, 0533)
	if err != nil {
		log.Panicf("Error while writing a file: %v", err)
	}
}

func countLeadingSpaces(line string) int {
	i := 0
	for _, runeValue := range line {
		if runeValue == ' ' {
			i++
		} else {
			break
		}
	}
	return i
}

func (fp FileProcessor) addNewRelease() string {
	var line, contentWithNewRelease string
	for _, char := range fp.finalContent {
		line += string(char)
		if char == '\n' {
			if strings.Contains(line, "- name: pre-rel"+strings.ReplaceAll(env.RefRelease, ".", "")) {
				releaseExtracts := fp.getNewReleaseExtracts(line)
				for _, re := range releaseExtracts {
					contentWithNewRelease += re
				}
			}
			contentWithNewRelease += line
			line = ""
		}
	}
	return contentWithNewRelease
}

func (fp FileProcessor) getNewReleaseExtracts(line string) []string {
	var releaseExtracts []string
	for _, sampleReleaseDef := range fp.sampleReleaseDefs {
		targetRelDefLine := ""
		for _, char := range sampleReleaseDef {
			targetRelDefLine += string(char)
			if char == '\n' {
				if strings.Contains(line, targetRelDefLine) {
					newDefs := fp.replaceRelStrings(sampleReleaseDef)
					for _, newDef := range newDefs {
						releaseExtracts = append(releaseExtracts, newDef)
					}
					return releaseExtracts
				}
			}
		}
	}
	return nil
}

func (fp FileProcessor) replaceRelStrings(sampleRelDef string) []string {
	var releaseExtracts []string
	for _, rel := range strings.Split(env.NewReleases, ",") {
		line := ""
		releaseExtract := ""
		for _, char := range sampleRelDef {
			line += string(char)
			if char == '\n' {
				line = strings.ReplaceAll(line, env.RefRelease, rel)
				line = strings.ReplaceAll(line, strings.ReplaceAll(env.RefRelease, ".", ""), strings.ReplaceAll(rel, ".", ""))
				releaseExtract += line
				line = ""
			}
		}
		releaseExtracts = append(releaseExtracts, releaseExtract)
		releaseExtract = ""
	}
	return releaseExtracts
}
