package main

import (
	"fmt"
	"os"
	"strings"

	imagesyncer "github.com/kyma-project/test-infra/pkg/imagesync"

	"gopkg.in/yaml.v3"
)

func getTarget(source, targetRepo, targetTag string) (string, error) {
	sourceParts := strings.Split(source, "/")
	repoName := sourceParts[len(sourceParts)-1] // Get the last part which should be repo:tag or repo@sha256

	// Add "library/" if the source image is not namespaced
	if len(sourceParts) == 1 {
		repoName = "library/" + repoName
	} else {
		repoName = strings.Join(sourceParts[:len(sourceParts)-1], "/") + "/" + repoName
	}

	target := targetRepo + repoName
	if strings.Contains(repoName, "@sha256:") {
		if targetTag == "" {
			return "", fmt.Errorf("sha256 digest detected, but the \"tag\" was not specified")
		}
		imageName := strings.Split(repoName, "@sha256:")[0]
		target = targetRepo + imageName + ":" + targetTag
		// Allow retagging when the source image is not using SHA256 hash
	} else if targetTag != "" {
		imageName := strings.Split(repoName, ":")[0]
		target = targetRepo + imageName + ":" + targetTag
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
	return &syncDef, nil
}
