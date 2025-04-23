package imagebuilder

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// reportRegex is a regular expression that matches the image build report
var (
	reportRegex = regexp.MustCompile(`(?s)---IMAGE BUILD REPORT---(.*)---END OF IMAGE BUILD REPORT---`)

	timestampRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z\s`)
)

type BuildReport struct {
	// Status is the overall status of the build including signing and pushing
	Status string `json:"status"`
	// IsPushed indicates whether the image was pushed to a registry
	IsPushed bool `json:"pushed"`
	// IsSigned indicates whether the image was signed
	IsSigned bool `json:"signed"`
	// Name is the name of the image
	Name string `json:"image_name"`
	// Images is a list of all built images
	Images []string `json:"images_list"`
	// Digest is the digest of the image
	Digest string `json:"digest"`
	// Tags is a list of tags for the image
	Tags []string `json:"tags"`
	// RegistryURL is the URL of the registry where the image was pushed
	RegistryURL string `json:"repository_path"`
	// Architecture is the architecture of the image
	Architecture []string `json:"architecture"`
}

func NewBuildReportFromLogs(log string) (*BuildReport, error) {
	// Strip all timestamps from log
	log = timestampRegex.ReplaceAllString(log, "")

	// Find the report in the log
	matches := reportRegex.FindStringSubmatch(log)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to find image build report in log")
	}

	// Parse the report data
	var report BuildReport
	if err := json.Unmarshal([]byte(matches[1]), &report); err != nil {
		return nil, err
	}

	return &report, nil
}

func WriteReportToFile(report *BuildReport, path string) error {
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	err = os.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write report to file: %w", err)
	}

	return nil
}
