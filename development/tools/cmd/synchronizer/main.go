package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	sc "github.com/kyma-project/test-infra/development/tools/cmd/synchronizer/syncomponent"
	t "github.com/kyma-project/test-infra/development/tools/cmd/synchronizer/tools"
	"github.com/pkg/errors"
)

const (
	envKymaProjectDir  = "KYMA_PROJECT_DIR"
	slackClientToken   = "SLACK_CLIENT_TOKEN"
	slackClientChannel = "STABILITY_SLACK_CLIENT_CHANNEL_ID"
	outOfDate          = "OUT_OF_DATE_DAYS"
	pathToVersionFile  = "Makefile"
	varsionPathCommand = "version-path"
)

// ComponentStorage inludes pack of components
type ComponentStorage struct {
	components []*sc.Component
}

// AddComponent adds new Component to storage
func (cs *ComponentStorage) AddComponent(comp *sc.Component) {
	cs.components = append(cs.components, comp)
}

func main() {
	rootDir := os.Getenv(envKymaProjectDir)
	slackToken := os.Getenv(slackClientToken)
	slackChannel := os.Getenv(slackClientChannel)
	outOfDateDays, err := strconv.Atoi(os.Getenv(outOfDate))

	if err != nil {
		log.Printf("Cannot tranform %s env to integer, default value will be used", outOfDate)
		outOfDateDays = 0
	}
	if rootDir == "" {
		log.Fatalf("missing env: %s", envKymaProjectDir)
	}
	if slackToken == "" {
		log.Fatalf("missing env: %s", slackClientToken)
	}
	if slackChannel == "" {
		log.Fatalf("missing env: %s", slackClientChannel)
	}

	storage := &ComponentStorage{}

	findComponents(rootDir, storage)
	fillComponentStorage(rootDir, storage, outOfDateDays)

	reports := sc.GenerateMessage(storage.components)
	alertAmount := len(reports)
	log.Printf("There are %d components with alerts \n", alertAmount)
	if alertAmount == 0 {
		return
	}

	messages := make([]t.Message, alertAmount)
	for i := range reports {
		messages[i] = reports[i]
	}
	err = t.SendMessage(slackToken, strings.Trim(slackChannel, "#"), messages)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func findComponents(dir string, storage *ComponentStorage) {
	runner := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip walk if the file is dir or name is not equal to 'pathToVersionFile' value
		if info.IsDir() || info.Name() != pathToVersionFile {
			return nil
		}
		// skip walk if path has `vendor` dir inside
		if strings.Contains(path, "/vendor/") {
			log.Printf("Path %q was skipped beacuse it is 'vendor' path", path)
			return nil
		}

		componentDir, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			return errors.Wrap(err, "while finding directory for pathToVersionFile")
		}

		versionPaths, err := findVersionPath(componentDir)
		if err != nil {
			return errors.Wrap(err, "while finding path to version")
		}
		// skip walk if path to values.yaml is empty, component has no info about version
		if len(versionPaths) == 0 {
			log.Printf("File %s in %q has no %q command", pathToVersionFile, path, varsionPathCommand)
			return nil
		}

		synComponent := sc.NewSynComponent(sc.RelativePathToComponent(dir, componentDir), versionPaths)
		storage.AddComponent(synComponent)
		return nil
	}

	if err := filepath.Walk(dir, runner); err != nil {
		log.Fatalf("Cannot walk for %q directory: %s", dir, err)
	}
}

func fillComponentStorage(dir string, storage *ComponentStorage, expiryDays int) {
	for _, component := range storage.components {
		// find component git hash
		hash, err := t.FindCommitHash(dir, component.Path)
		if err != nil {
			log.Fatal(err.Error())
		}
		component.GitHash = hash

		// find component date commit
		date, err := t.FindCommitDate(dir, component.Path)
		if err != nil {
			log.Fatal(err.Error())
		}
		component.CommitDate = date

		// find component versions
		err = t.FindComponentVersion(dir, component)
		if err != nil {
			log.Fatal(err.Error())
		}

		// find modified files beetwen versions and hash
		for _, version := range component.Versions {
			files, err := t.FindFileDifference(dir, component.Path, component.GitHash, version.Version)
			if err != nil {
				log.Fatal(err.Error())
			}
			version.ModifiedFiles = files
		}
		component.SetOutOfDate(expiryDays)
	}
}

// fetch data about paths to values.yaml files where version of component exist
func findVersionPath(dir string) ([]string, error) {
	cmd := exec.Command("bash", "-c", "make "+varsionPathCommand)
	cmd.Dir = dir

	res, err := cmd.CombinedOutput()
	response := string(res)
	if err == nil {
		response = strings.Trim(response, "\n")
		return strings.Split(response, "\n"), nil
	}

	re := regexp.MustCompile("No rule to make target .*" + varsionPathCommand + ".*")
	if re.MatchString(string(response)) {
		return []string{}, nil
	}

	return []string{}, err
}
