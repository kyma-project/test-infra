package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kyma-project/test-infra/development/image-detector/internal/prowjob"
	"github.com/kyma-project/test-infra/development/image-detector/internal/terraform"
	"gopkg.in/yaml.v3"
	"k8s.io/test-infra/prow/config"
)

type options struct {
	ProwConfig    string
	JobsConfigDir string
	TerraformDir  string
}

func (o *options) loadOptions(flagSet *flag.FlagSet) *flag.FlagSet {
	flagSet.StringVar(&o.ProwConfig, "prow-config", "", "path to the prow config file")
	flagSet.StringVar(&o.JobsConfigDir, "job-config-dir", "", "path to directory containing prow jobs definition")
	flagSet.StringVar(&o.TerraformDir, "terraform-dir", "", "path to directory containing terraform files")
	return flagSet
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

	// load images from security-config

	var images []string
	// Ignore prow jobs if paths are not provided
	if o.ProwConfig != "" && o.JobsConfigDir != "" {
		cfg, err := loadJobConfigs(o)
		if err != nil {
			log.Fatalf("Failed to load prow job config: %s", err)
		}

		images = append(images, prowjob.ExtractFromJobConfig(cfg.JobConfig)...)
	}

	// Ignore terraform if path not provided
	if o.TerraformDir != "" {
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
	}

	images = uniqueImages(images)

	// Write images to security-config
	writeImagesToFile("security-config.yaml", images)
}

func loadJobConfigs(o options) (*config.Config, error) {
	cfg, err := config.Load(o.ProwConfig, o.JobsConfigDir, nil, "")
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func writeImagesToFile(path string, images []string) error {
	fileData := ImagesFile{
		Images: images,
	}

	encoded, err := yaml.Marshal(fileData)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	_, err = f.Write(encoded)
	if err != nil {
		return err
	}

	return nil
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
