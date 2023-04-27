package main

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

func TestLoadOptions(t *testing.T) {
	tc := []struct {
		Name            string
		Flags           []string
		WantErr         bool
		ExpectedOptions options
	}{
		{
			Name:            "unknown flag, fail",
			WantErr:         true,
			Flags:           []string{"-unknown=tets"},
			ExpectedOptions: options{},
		},
		{
			Name:    "all flags, pass",
			WantErr: false,
			Flags: []string{
				"-prow-config=../../prow/config.yaml",
				"-job-config-dir=../../prow/jobs",
				"-terraform-dir=../../configs/terraform",
				"-sec-scanner-config=./sec-scanner-config.yaml",
			},
			ExpectedOptions: options{
				ProwConfig:       "../../prow/config.yaml",
				JobsConfigDir:    "../../prow/jobs",
				TerraformDir:     "../../configs/terraform",
				SecScannerConfig: "./sec-scanner-config.yaml",
			},
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			o := options{}
			o.loadOptions(flagSet)
			err := flagSet.Parse(c.Flags)
			if err != nil && !c.WantErr {
				t.Errorf("unexcpected error occurred: %s", err)
			}

			if !reflect.DeepEqual(o, c.ExpectedOptions) {
				t.Errorf("%v != %v", o, c.ExpectedOptions)
			}
		})
	}
}

func TestValidateOptions(t *testing.T) {
	tc := []struct {
		Name    string
		WantErr bool
		Options options
	}{
		{
			Name:    "empty options, fail",
			WantErr: true,
			Options: options{},
		},
		{
			Name:    "missing prow config path, fail",
			WantErr: true,
			Options: options{
				JobsConfigDir:    "prow/jobs",
				TerraformDir:     "terraform",
				SecScannerConfig: "sec-scanner-config.yaml",
				KubernetesFiles:  "kubernetes",
			},
		},
		{
			Name:    "missing prow jobs directory path, fail",
			WantErr: true,
			Options: options{
				ProwConfig:       "prow/config.yaml",
				TerraformDir:     "terraform",
				SecScannerConfig: "sec-scanner-config.yaml",
				KubernetesFiles:  "kubernetes",
			},
		},
		{
			Name:    "missing terraform directory path, fail",
			WantErr: true,
			Options: options{
				ProwConfig:       "prow/config.yaml",
				JobsConfigDir:    "prow/jobs",
				SecScannerConfig: "sec-scanner-config.yaml",
				KubernetesFiles:  "kubernetes",
			},
		},
		{
			Name:    "missing security scanner config path, fail",
			WantErr: true,
			Options: options{
				ProwConfig:      "prow/config.yaml",
				JobsConfigDir:   "prow/jobs",
				TerraformDir:    "terraform",
				KubernetesFiles: "kubernetes",
			},
		},
		{
			Name:    "missing kubernetes directory path, fail",
			WantErr: true,
			Options: options{
				ProwConfig:       "prow/config.yaml",
				JobsConfigDir:    "prow/jobs",
				TerraformDir:     "terraform",
				SecScannerConfig: "sec-scanner-config.yaml",
			},
		},
		{
			Name:    "all provided, pass",
			WantErr: false,
			Options: options{
				ProwConfig:       "prow/config.yaml",
				JobsConfigDir:    "prow/jobs",
				TerraformDir:     "terraform",
				SecScannerConfig: "sec-scanner-config.yaml",
				KubernetesFiles:  "kubernetes",
			},
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			o := c.Options
			err := o.validateOptions()
			if !c.WantErr && err != nil {
				t.Errorf("error occurred, but unexpected: %s", err)
			}

			if c.WantErr && err == nil {
				t.Errorf("error not occurred, but expected")
			}
		})
	}
}

func TestUniqueImages(t *testing.T) {
	tc := []struct {
		Name           string
		GivenImages    []string
		ExpectedImages []string
	}{
		{
			Name:           "remove duplicated images",
			GivenImages:    []string{"same/image:test", "same/image:test"},
			ExpectedImages: []string{"same/image:test"},
		},
		{
			Name:           "keep images for unique list",
			GivenImages:    []string{"unique/image:123", "other-unique/image:123"},
			ExpectedImages: []string{"unique/image:123", "other-unique/image:123"},
		},
		{
			Name:           "find multiple duplicates in long images list",
			GivenImages:    []string{"some/image:test", "other/image:test", "other/image:test", "some-other/image:test", "one-more/image:test-tag", "yet/another-image:other-tag", "yet/another-image:other-tag", "yet/another-image:tag", "yet/another-image:other-tag"},
			ExpectedImages: []string{"some/image:test", "other/image:test", "some-other/image:test", "one-more/image:test-tag", "yet/another-image:other-tag", "yet/another-image:tag"},
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			actual := uniqueImages(c.GivenImages)

			if !reflect.DeepEqual(actual, c.ExpectedImages) {
				t.Errorf("%v != %v", actual, c.ExpectedImages)
			}
		})
	}
}
