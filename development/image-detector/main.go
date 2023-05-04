package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kyma-project/test-infra/development/pkg/extractimages"
	"github.com/kyma-project/test-infra/development/pkg/securityconfig"
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

	images = append(images, extractimages.FromProwJobConfig(cfg.JobConfig)...)

	// terraform
	files, err := findFilesInDirectory(o.TerraformDir, ".*.(tf|tfvars)")
	if err != nil {
		log.Fatalf("failed to find files in terraform directory %s: %s", o.TerraformDir, err)
	}

	img, err := extractImagesFromFiles(files, extractimages.FromTerraform)
	if err != nil {
		log.Fatalf("failed to extract images from terraform files: %s", err)
	}

	images = append(images, img...)

	// kubernetes
	files, err = findFilesInDirectory(o.KubernetesFiles, ".*.(yaml|yml)")
	if err != nil {
		log.Fatalf("failed to find files in kubernetes directory %s: %s", o.KubernetesFiles, err)
	}

	img, err = extractImagesFromFiles(files, extractimages.FromKubernetesDeployments)
	if err != nil {
		log.Fatalf("failed to extract images from kubernetes files: %s", err)
	}

	images = append(images, img...)

	images = uniqueImages(images)

	// Write images to security-config
	config.Images = images
	config.SaveToFile(o.SecScannerConfig)
}

func extractImagesFromFiles(files []string, extract func(path string) ([]string, error)) ([]string, error) {
	var images []string
	for _, file := range files {
		img, err := extract(file)
		if err != nil {
			return nil, fmt.Errorf("failed to extract images from file %s: %s", file, err)
		}

		images = append(images, img...)
	}

	return images, nil
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
