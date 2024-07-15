package main

import (
	"fmt"
	"os"
	"strings"

	imagesyncer "github.com/kyma-project/test-infra/pkg/imagesync"

	"gopkg.in/yaml.v3"
)

func getTarget(source, targetRepo, targetTag string) (string, error) {
	// Parse the source image to extract the repository name
	sourceParts := strings.Split(source, "/")
	repoName := sourceParts[len(sourceParts)-1] // Get the last part which should be repo:tag or repo@sha256

	if strings.Contains(repoName, ":") {
		repoName = strings.Split(repoName, ":")[0]
	} else if strings.Contains(repoName, "@sha256:") {
		repoName = strings.Split(repoName, "@sha256:")[0]
	}

	target := fmt.Sprintf("%s%s", targetRepo, repoName)
	if strings.Contains(source, "@sha256:") {
		if targetTag == "" {
			return "", fmt.Errorf("sha256 digest detected, but the \"tag\" was not specified")
		}
		target = fmt.Sprintf("%s:%s", target, targetTag)
	} else if targetTag != "" {
		target = fmt.Sprintf("%s:%s", target, targetTag)
	}

	return target, nil
}

func parseImagesFile(file string) (*imagesyncer.SyncDef, error) {
	f, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var syncDef imagesyncer.SyncDef
	if err := yaml.Unmarshal(f, &syncDef); err != nil {
		return nil, err
	}
	if syncDef.TargetRepoPrefix == "" {
		return nil, fmt.Errorf("targetRepoPrefix can not be empty")
	}
	return &syncDef, nil
}
