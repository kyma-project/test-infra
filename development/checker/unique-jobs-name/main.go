package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/config"
)

type options struct {
	prowConfig   string
	jobConfigDir string
}

type report struct {
	foundDuplicates bool
	messages        []string
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

	if err != nil {
		logrus.Fatalf("Error during fetching jobs files: %s", err)
	}

	jobs := map[string]int{}

	c, err := config.Load(o.prowConfig, o.jobConfigDir, nil, "")
	if err != nil {
		logrus.Fatalf("Cannot load config from directory %q: %s", o.jobConfigDir, err)
	}

	addPreSubmitJobsName(jobs, c.JobConfig.PresubmitsStatic)
	addPostSubmitJobsName(jobs, c.JobConfig.PostsubmitsStatic)

	rep := report{foundDuplicates: false}
	for name, val := range jobs {
		if val > 1 {
			rep.foundDuplicates = true
			rep.messages = append(rep.messages, fmt.Sprintf("Prow job %q has %d instances", name, val))
		}
	}

	if rep.foundDuplicates {
		logrus.Fatalf("Config jobs are not unique:\n%s", rep.toString())
	}
}

func addPreSubmitJobsName(all map[string]int, config map[string][]config.Presubmit) {
	for _, jobKind := range config {
		for _, pre := range jobKind {
			n := pre.Name
			curr := all[n]
			all[n] = curr + 1
		}
	}
}

func addPostSubmitJobsName(all map[string]int, config map[string][]config.Postsubmit) {
	for _, jobKind := range config {
		for _, post := range jobKind {
			n := post.Name
			curr := all[n]
			all[n] = curr + 1
		}
	}
}
