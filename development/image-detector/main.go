package main

import (
	"context"
	"log"
	"os"

	"github.com/kyma-project/test-infra/development/image-detector/bumper"
	"github.com/kyma-project/test-infra/development/pkg/extractimageurls"
	"github.com/kyma-project/test-infra/development/pkg/securityconfig"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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

	// TektonCatalog contains root path to tekton catalog directory
	TektonCatalog string

	// AutobumpConfig contains root path to config for autobumper for sec-scanner-config
	AutobumpConfig string
)

var rootCmd = &cobra.Command{
	Use:   "image-detector",
	Short: "Image Detector CLI",
	Long:  "Command-Line tool to retrieve list of images and update security-config",
	Run: func(cmd *cobra.Command, args []string) {
		// load images from security config
		reader, err := os.Open(SecScannerConfig)
		if err != nil {
			log.Fatalf("failed to open security config file %s", err)
		}
		securityConfig, err := securityconfig.ParseSecurityConfig(reader)
		if err != nil {
			log.Fatalf("failed to parse security config file: %s", err)
		}

		images := securityConfig.Images

		// get images from prow jobs
		if ProwConfig != "" && JobsConfigDir != "" {
			prowConfig, err := config.Load(ProwConfig, JobsConfigDir, nil, "")
			if err != nil {
				log.Fatalf("failed to load prow job config: %s", err)
			}

			images = append(images, extractimageurls.FromProwJobConfig(prowConfig.JobConfig)...)
		}

		// get images from terraform
		if TerraformDir != "" {
			files, err := extractimageurls.FindFilesInDirectory(TerraformDir, ".*.(tf|tfvars)")
			if err != nil {
				log.Fatalf("failed to find files in terraform directory %s: %s", TerraformDir, err)
			}

			imgs, err := extractimageurls.FromFiles(files, extractimageurls.FromTerraform)
			if err != nil {
				log.Fatalf("failed to extract images from terraform files: %s", err)
			}

			images = append(images, imgs...)
		}

		// get images from kubernetes
		if KubernetesFiles != "" {
			files, err := extractimageurls.FindFilesInDirectory(KubernetesFiles, ".*.(yaml|yml)")
			if err != nil {
				log.Fatalf("failed to find files in kubernetes directory %s: %s", KubernetesFiles, err)
			}

			imgs, err := extractimageurls.FromFiles(files, extractimageurls.FromKubernetesDeployments)
			if err != nil {
				log.Fatalf("failed to extract images from kubernetes files: %s", err)
			}

			images = append(images, imgs...)
		}

		// get images from tekton catalog
		if TektonCatalog != "" {
			files, err := extractimageurls.FindFilesInDirectory(TektonCatalog, ".*.(yaml|yml)")
			if err != nil {
				log.Fatalf("failed to find files in tekton catalog directory %s: %s", TektonCatalog, err)
			}

			imgs, err := extractimageurls.FromFiles(files, extractimageurls.FromTektonTask)
			if err != nil {
				log.Fatalf("failed to extract image urls from tekton tasks files: %s", err)
			}

			images = append(images, imgs...)
		}

		images = extractimageurls.UniqueImages(images)

		// write images to security config
		securityConfig.Images = images
		securityConfig.SaveToFile(SecScannerConfig)

		// Run autbumper if autobump config provided
		if AutobumpConfig != "" {
			err := runAutobumper(AutobumpConfig)
			if err != nil {
				log.Fatalf("failed to run bumper: %s", err)
			}
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&ProwConfig, "prow-config", "", "path to the Prow config file")
	rootCmd.PersistentFlags().StringVar(&JobsConfigDir, "prow-jobs-dir", "", "path to the directory which contains Prow job files")
	rootCmd.PersistentFlags().StringVar(&TerraformDir, "terraform-dir", "", "path to the directory containing Terraform files")
	rootCmd.PersistentFlags().StringVar(&SecScannerConfig, "sec-scanner-config", "", "path to the security scanner config field")
	rootCmd.PersistentFlags().StringVar(&KubernetesFiles, "kubernetes-dir", "", "path to the directory containing Kubernetes deployments")
	rootCmd.PersistentFlags().StringVar(&TektonCatalog, "tekton-catalog", "", "path to the Tekton catalog directory")
	rootCmd.PersistentFlags().StringVar(&AutobumpConfig, "autobump-config", "", "path to the config for autobumper for security scanner config")

	rootCmd.MarkFlagRequired("sec-scanner-config")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to run command: %s", err)
	}
}

// client is bumper client
type client struct {
	o *options
}

// Changes returns a slice of functions, each one does some stuff, and
// returns commit message for the changes
func (c *client) Changes() []func(context.Context) (string, error) {
	return []func(context.Context) (string, error){
		func(ctx context.Context) (string, error) {
			return "Bumping sec-scanner-config.yml", nil
		},
	}
}

// PRTitleBody returns the body of the PR, this function runs after each commit
func (c *client) PRTitleBody() (string, string, error) {
	return "Update sec-scanner-config.yml" + "\n", "", nil
}

// options is the options for autobumper operations.
type options struct {
	GitHubRepo      string   `yaml:"gitHubRepo"`
	FoldersToFilter []string `yaml:"foldersToFilter"`
	FilesToFilter   []string `yaml:"filesToFilter"`
}

// runAutobumper is wrapper for bumper API -> ACL
func runAutobumper(autoBumperCfg string) error {
	f, err := os.Open(autoBumperCfg)
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(f)

	var bumperClientOpt options
	err = decoder.Decode(&bumperClientOpt)
	if err != nil {
		return err
	}

	var opts bumper.Options
	err = decoder.Decode(&opts)
	if err != nil {
		return err
	}

	ctx := context.Background()
	bumper.Run(ctx, &opts, &client{o: &bumperClientOpt})

	return nil
}
