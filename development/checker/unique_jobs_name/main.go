package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"

	"os"
	"path/filepath"
)

type options struct {
	prowConfig   string
	jobConfigDir string
}

type report struct {
	active   bool
	messages []string
}

func (r report) toString() string {
	response := ""
	for _, message := range r.messages {
		response += message + "\n"
	}
	return response
}

func validationOptions() (options, error) {
	o := options{}

	flag.StringVar(&o.prowConfig, "config-path", "", "Path to config file.")
	flag.StringVar(&o.jobConfigDir, "jobs-config-dir", "", "Path to dir with prow job configs.")
	flag.Parse()

	if o.prowConfig == "" {
		return o, errors.New("Path to prow config file is required")
	}
	if o.jobConfigDir == "" {
		return o, errors.New("Path to dir with job config files is required")
	}

	return o, nil
}

func main() {
	o, err := validationOptions()
	if err != nil {
		logrus.Fatalf("Error during reads flags: %s", err)
	}

	if _, err := os.Stat(o.prowConfig); os.IsNotExist(err) {
		logrus.Fatalf("Cannot find prow config file: %s", err)
	}

	jobConfigFiles, err := findJobsConfigFiles(o.jobConfigDir)
	if err != nil {
		logrus.Fatalf("Error during fetching jobs files: %s", err)
	}

	jobs := map[string]int{}

	for _, file := range jobConfigFiles {
		c, err := config.Load(o.prowConfig, file)
		if err != nil {
			logrus.Fatalf("Cannot load config from file %q: %s", file, err)
		}

		addPreSubmitJobsName(jobs, c.Presubmits)
		addPostSubmitJobsName(jobs, c.Postsubmits)
	}

	rep := report{active: false}
	for name, val := range jobs {
		if val > 1 {
			rep.active = true
			rep.messages = append(rep.messages, fmt.Sprintf("Prow job %q has %d instances", name, val))
		}
	}

	if rep.active {
		logrus.Fatalf("Config jobs are not unique:\n%s", rep.toString())
	}
}

func findJobsConfigFiles(dir string) ([]string, error) {
	out := []string{}

	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(info.Name()) != ".yaml" {
				return nil
			}
			out = append(out, path)
			return nil
		})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func addPreSubmitJobsName(jobs map[string]int, c map[string][]config.Presubmit) {
	for _, jobKind := range c {
		for _, pre := range jobKind {
			n := pre.Name
			curr := jobs[n]
			jobs[n] = curr + 1
		}
	}
}

func addPostSubmitJobsName(jobs map[string]int, c map[string][]config.Postsubmit) {
	for _, jobKind := range c {
		for _, post := range jobKind {
			n := post.Name
			curr := jobs[n]
			jobs[n] = curr + 1
		}
	}
}
