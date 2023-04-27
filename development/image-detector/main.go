package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kyma-project/test-infra/development/image-detector/internal/prowjob"
	"github.com/kyma-project/test-infra/development/image-detector/internal/securityconfig"
	"github.com/kyma-project/test-infra/development/image-detector/internal/terraform"
	"k8s.io/test-infra/prow/config"
)

type options struct {
	ProwConfig       string
	JobsConfigDir    string
	TerraformDir     string
	SecScannerConfig string
	KubernetesFiles  string
}

func (o *options) loadOptions(flagSet *flag.FlagSet) *flag.FlagSet {
	flagSet.StringVar(&o.ProwConfig, "prow-config", "", "path to the prow config file")
	flagSet.StringVar(&o.JobsConfigDir, "job-config-dir", "", "path to directory containing prow jobs definition")
	flagSet.StringVar(&o.TerraformDir, "terraform-dir", "", "path to directory containing terraform files")
	flagSet.StringVar(&o.SecScannerConfig, "sec-scanner-config", "", "path to security config file")
	flagSet.StringVar(&o.KubernetesFiles, "kubernetes-dir", "", "path to kubernetes deployments")
	return flagSet
}

func (o *options) validateOptions() error {
	if o.ProwConfig == "" {
		return fmt.Errorf("prow config path must be provided")
	}

	if o.JobsConfigDir == "" {
		return fmt.Errorf("job config directory path must be provided")
	}

	if o.TerraformDir == "" {
		return fmt.Errorf("terraform directory path must be provided")
	}

	if o.SecScannerConfig == "" {
		return fmt.Errorf("security scanners config file path must be provided")
	}

	if o.KubernetesFiles == "" {
		return fmt.Errorf("kubernetes files path must be provided")
	}

	return nil
}

type Detector interface {
	// Extract list of images used in given file
	Extract(path string) ([]string, error)
}

type ImagesFile struct {
	Images []string `yaml:"images"`
}

func main() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o := options{}
	o.loadOptions(flagSet)
	flagSet.Parse(os.Args[1:])
	err := o.validateOptions()
	if err != nil {
		log.Fatalf("provided options are invalid: %s", err)
	}

	// load images from security-config
	config, err := securityconfig.LoadSecurityConfig(o.SecScannerConfig)
	if err != nil {
		log.Fatalf("failed to load security config file: %s", err)
	}

	// prow jobs
	images := config.Images
	cfg, err := loadJobConfigs(o)
	if err != nil {
		log.Fatalf("Failed to load prow job config: %s", err)
	}

	images = append(images, prowjob.ExtractFromJobConfig(cfg.JobConfig)...)

	// terraform
	files, err := findFilesInDirectory(o.TerraformDir, ".*.(tf|tfvars)")
	if err != nil {
		log.Fatalf("failed to find files in terraform directory %s: %s", o.TerraformDir, err)
	}

	for _, file := range files {
		imgs, err := terraform.Extract(file)
		if err != nil {
			log.Fatalf("failed to extract images from terraform: %s", err)
		}

		images = append(images, imgs...)
	}

	images = uniqueImages(images)

	// Write images to security-config
	config.Images = images
	config.SaveToFile(o.SecScannerConfig)
}

func loadJobConfigs(o options) (*config.Config, error) {
	cfg, err := config.Load(o.ProwConfig, o.JobsConfigDir, nil, "")
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func findFilesInDirectory(rootPath, regex string) ([]string, error) {
	var files []string
	err := filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		re := regexp.MustCompile(regex)

		if re.MatchString(path) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func uniqueImages(images []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range images {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}
