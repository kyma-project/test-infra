package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	sc "github.com/kyma-project/test-infra/development/tools/pkg/synchronizer/syncomponent"
	t "github.com/kyma-project/test-infra/development/tools/pkg/synchronizer/tools"
	"github.com/pkg/errors"
)

const (
	envKymaProjectDir     = "KYMA_PROJECT_DIR"
	envSlackClientToken   = "SLACK_CLIENT_TOKEN"
	envSlackClientChannel = "STABILITY_SLACK_CLIENT_CHANNEL_ID"
	envOutOfDateThreshold = "OUT_OF_DATE_DAYS"
	defaultOutOfDateDays  = 3
	pathToVersionFile     = "Makefile"
	varsionPathCommand    = "version-path"
)

// ComponentStorage includes pack of components
type ComponentStorage struct {
	components []*sc.Component
}

// AddComponent adds new Component to storage
func (cs *ComponentStorage) AddComponent(comp *sc.Component) {
	cs.components = append(cs.components, comp)
}

func main() {
	rootDir := os.Getenv(envKymaProjectDir)
	slackToken := os.Getenv(envSlackClientToken)
	slackChannel := os.Getenv(envSlackClientChannel)
	outOfDateDays, err := strconv.Atoi(os.Getenv(envOutOfDateThreshold))

	if err != nil {
		log.Printf("Cannot tranform %s env to integer, default value '%d' will be used", envOutOfDateThreshold, defaultOutOfDateDays)
		outOfDateDays = defaultOutOfDateDays
	}
	if rootDir == "" {
		log.Fatalf("missing env: %s", envKymaProjectDir)
	}
	sendMessageToSlack := true
	if slackToken == "" {
		sendMessageToSlack = false
		log.Printf("missing env: %s, alert message will not be sent to slack", envSlackClientToken)
	}
	if slackChannel == "" {
		sendMessageToSlack = false
		log.Printf("missing env: %s, alert message will not be sent to slack", envSlackClientChannel)
	}

	storage, err := generateComponentStorage(rootDir)
	if err != nil {
		log.Fatalf("Cannot generate component storage: %s", err.Error())
	}
	fillComponentStorage(rootDir, storage, outOfDateDays)

	reports := sc.GenerateReport(storage.components)
	alertAmount := len(reports)
	log.Printf("There are %d components with alerts \n", alertAmount)
	if alertAmount == 0 {
		return
	}

	for _, report := range reports {
		log.Printf("Component %q is out of date: \n%s \n", report.GetTitle(), report.GetValue())
	}

	if sendMessageToSlack {
		messages := make([]t.Message, alertAmount)
		for i := range reports {
			messages[i] = reports[i]
		}
		err = t.SendMessage(slackToken, strings.Trim(slackChannel, "#"), messages)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func generateComponentStorage(dir string) (*ComponentStorage, error) {
	storage := &ComponentStorage{}

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

		path, err = filepath.Rel(dir, componentDir)
		if err != nil {
			return errors.Wrapf(err, "while trying get relative path from %q", componentDir)
		}

		synComponent := sc.NewSynComponent(path, versionPaths)
		storage.AddComponent(synComponent)
		return nil
	}

	if err := filepath.Walk(dir, runner); err != nil {
		return nil, errors.Wrapf(err, "while walking for %q directory: %s")
	}

	return storage, nil
}

func fillComponentStorage(dir string, storage *ComponentStorage, expiryDays int) {
	for _, component := range storage.components {
		// find component git hash
		hash, err := t.FindCommitHash(dir, component.Path)
		if err != nil {
			log.Fatal(err.Error())
		}
		component.GitHash = hash

		// find component git hashes and commits date
		gitHistory, err := t.FetchCommitsHistory(dir, component.Path, expiryDays)
		if err != nil {
			log.Fatal(err.Error())
		}
		component.GitHashHistory = gitHistory

		// find component versions
		err = t.FindComponentVersion(dir, component)
		if err != nil {
			log.Fatal(err.Error())
		}

		// find modified files beetwen versions and the oldest allowed hash
		for _, version := range component.Versions {
			files, err := t.FindFileDifference(dir, component.Path, component.GetOldest(), version.Version)
			if err != nil {
				log.Fatal(err.Error())
			}
			version.ModifiedFiles = files
		}

		component.CheckIsOutOfDate()
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
