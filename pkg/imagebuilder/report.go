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
	// IsProduction indicates whether the image is a production image
	IsProduction bool `json:"is_production"`
	// ImageSpec contains the image name, tags, and repository path
	ImageSpec ImageSpec `json:"image_spec"`
}

type ImageSpec struct {
	Name           string   `json:"image_name"`
	Tags           []string `json:"tags"`
	RepositoryPath string   `json:"repository_path"`
}

func (br *BuildReport) GetImages() []string {
	var images []string

	if br == nil {
		return images
	}

	for _, tag := range br.ImageSpec.Tags {
		images = append(images, fmt.Sprintf("%s%s:%s", br.ImageSpec.RepositoryPath, br.ImageSpec.Name, tag))
	}

	return images
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
