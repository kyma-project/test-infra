package main

import (
	"fmt"
	"os"
	"strings"

	imagesyncer "github.com/kyma-project/test-infra/development/image-syncer/pkg"
	"gopkg.in/yaml.v2"
)

func getTarget(source, targetRepo, targetTag string) (string, error) {
	target := targetRepo + source
	if strings.Contains(source, "@sha256:") {
		if targetTag == "" {
			return "", fmt.Errorf("sha256 digest detected, but the \"tag\" was not specified")
		}
		imageName := strings.Split(source, "@sha256:")[0]
		target = targetRepo + imageName + ":" + targetTag
		// Allow retagging when the source image is not using SHA256 hash
	} else if targetTag != "" {
		imageName := strings.Split(source, ":")[0]
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
	if syncDef.TargetRepoPrefix == "" {
		return nil, fmt.Errorf("targetRepoPrefix can not be empty")
	}
	return &syncDef, nil
}
