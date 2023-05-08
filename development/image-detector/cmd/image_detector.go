package cmd

import (
	"log"

	"github.com/kyma-project/test-infra/development/image-detector/pkg"
	"github.com/kyma-project/test-infra/development/pkg/extractimages"
	"github.com/kyma-project/test-infra/development/pkg/securityconfig"
	"github.com/spf13/cobra"
	"k8s.io/test-infra/prow/config"
)

var (
	// ProwConfig contains path to prow config file
	ProwConfig string

	// JobsConfigDir contains root path for directory containing Prow Jobs Configs
	JobsConfigDir string

	// TerraformDir contains root path to directory containing terraform files
	TerraformDir string

	// SecScannerConfig contains path to security scanners config .yaml file
	SecScannerConfig string

	// KubernetesFiles contains root path to directory containing kubernetes deployments file
	KubernetesFiles string
)

var rootCmd = &cobra.Command{
	Use:   "image-detector",
	Short: "Image Detector CLI",
	Long:  "Command-Line tool to retrieve list of images and update security-config",
	Run: func(cmd *cobra.Command, args []string) {
		// load images from security config
		securityConfig, err := securityconfig.LoadSecurityConfig(SecScannerConfig)
		if err != nil {
			log.Fatalf("failed to load security config file: %s", err)
		}

		images := securityConfig.Images

		// get images from prow jobs
		prowConfig, err := config.Load(ProwConfig, JobsConfigDir, nil, "")
		if err != nil {
			log.Fatalf("failed to load prow job config: %s", err)
		}

		images = append(images, extractimages.FromProwJobConfig(prowConfig.JobConfig)...)

		// get images from terraform
		files, err := pkg.FindFilesInDirectory(TerraformDir, ".*.(tf|tfvars)")
		if err != nil {
			log.Fatalf("failed to find files in terraform directory %s: %s", TerraformDir, err)
		}

		imgs, err := pkg.ExtractImagesFromFiles(files, extractimages.FromTerraform)
		if err != nil {
			log.Fatalf("failed to extract images from terraform files: %s", err)
		}

		images = append(images, imgs...)

		// get images from kubernetes
		files, err = pkg.FindFilesInDirectory(KubernetesFiles, ".*.(yaml|yml)")
		if err != nil {
			log.Fatalf("failed to find files in kubernetes directory %s: %s", KubernetesFiles, err)
		}

		imgs, err = pkg.ExtractImagesFromFiles(files, extractimages.FromKubernetesDeployments)
		if err != nil {
			log.Fatalf("failed to extract images from kubernetes files: %s", err)
		}

		images = append(images, imgs...)

		images = pkg.UniqueImages(images)

		// write images to security config
		securityConfig.Images = images
		securityConfig.SaveToFile(SecScannerConfig)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&ProwConfig, "prow-config", "", "path to the prow config file")
	rootCmd.PersistentFlags().StringVar(&JobsConfigDir, "prow-jobs-dir", "", "path to the directory which contains prow job files")
	rootCmd.PersistentFlags().StringVar(&TerraformDir, "terraform-dir", "", "path to the directory containing terraform files")
	rootCmd.PersistentFlags().StringVar(&SecScannerConfig, "sec-scanner-config", "", "path to the security scanner config fiel")
	rootCmd.PersistentFlags().StringVar(&KubernetesFiles, "kubernetes-dir", "", "path to the directory containing kubernetes deployments")

	rootCmd.MarkFlagRequired("prow-config")
	rootCmd.MarkFlagRequired("prow-jobs-dir")
	rootCmd.MarkFlagRequired("terraform-dir")
	rootCmd.MarkFlagRequired("sec-scanner-config")
	rootCmd.MarkFlagRequired("kubernetes-dir")
}
