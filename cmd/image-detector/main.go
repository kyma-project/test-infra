package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/kyma-project/test-infra/pkg/github/bumper"

	"github.com/kyma-project/test-infra/pkg/extractimageurls"
	"github.com/kyma-project/test-infra/pkg/securityconfig"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/prow/prow/config"
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

	// AutobumpConfig contains root path to config for autobumper for sec-scanners-config
	AutobumpConfig string

	// InRepoConfig contains path to the configuration of repositories with Prow inrepo config enabled
	InRepoConfig string

	// GithubTokenPath path to file containing github token for fetching inrepo config
	GithubTokenPath string
)

var rootCmd = &cobra.Command{
	Use:   "image-detector",
	Short: "Image Detector CLI",
	Long:  "Command-Line tool to retrieve list of images and update security-config",
	//nolint:revive
	Run: func(cmd *cobra.Command, args []string) {
		// load security config
		reader, err := os.Open(SecScannerConfig)
		if err != nil {
			log.Fatalf("failed to open security config file %s", err)
		}
		securityConfig, err := securityconfig.ParseSecurityConfig(reader)
		if err != nil {
			log.Fatalf("failed to parse security config file: %s", err)
		}

		// don't use previously scanned images as it will not delete removed once
		images := []string{}

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

		// get prow jobs configuration from in-repo configuration
		if InRepoConfig != "" {
			// load InRepo configuration
			file, err := os.Open(InRepoConfig)
			if err != nil {
				log.Fatalf("failed to load inrepo configuration: %s", err)
			}

			// parse configuration
			var cfg []extractimageurls.Repository
			err = yaml.NewDecoder(file).Decode(&cfg)
			if err != nil {
				log.Fatalf("failed to decode inrepo configuration: %s", err)
			}

			// load github token from env
			ghToken, err := loadGithubToken(GithubTokenPath)
			if err != nil {
				log.Fatalf("failed to load github token from %s: %s", GithubTokenPath, err)
			}

			for _, repo := range cfg {
				imgs, err := extractimageurls.FromInRepoConfig(repo, ghToken)
				if err != nil {
					log.Printf("warn: failed to extract image urls from repository %s: %v", &repo, err)
					continue
				}

				images = append(images, imgs...)
			}
		}

		images = extractimageurls.UniqueImages(images)

		// sort list of images to have consistent order
		sort.Strings(images)

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
	rootCmd.PersistentFlags().StringVar(&AutobumpConfig, "autobump-config", "", "path to the config for autobumper for security scanner config")
	rootCmd.PersistentFlags().StringVar(&InRepoConfig, "inrepo-config", "", "path to the configuration of repositories with Prow inrepo config enabled")
	rootCmd.PersistentFlags().StringVar(&GithubTokenPath, "github-token-path", "/etc/github/token", "path to github token for fetching inrepo config")

	rootCmd.MarkFlagRequired("sec-scanner-config")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to run command: %s", err)
	}
}

// loadGithubToken read github token from given file
func loadGithubToken(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// client is bumper client
type client struct {
	o *options
}

// Changes returns a slice of functions, each one does some stuff, and
// returns commit message for the changes
//
//nolint:revive
func (c *client) Changes() []func(context.Context) (string, []string, error) {
	return []func(context.Context) (string, []string, error){
		func(ctx context.Context) (string, []string, error) {
			return "Bumping sec-scanners-config.yaml", []string{"sec-scanners-config.yaml"}, nil
		},
	}
}

// PRTitleBody returns the body of the PR, this function runs after each commit
func (c *client) PRTitleBody() (string, string, error) {
	return "Update sec-scanners-config.yaml", "", nil
}

// options is the options for autobumper operations.
type options struct {
	GitHubRepo      string   `yaml:"gitHubRepo"`
	FoldersToFilter []string `yaml:"foldersToFilter"`
	FilesToFilter   []string `yaml:"filesToFilter"`
}

// runAutobumper is wrapper for bumper API -> ACL
func runAutobumper(autoBumperCfg string) error {
	data, err := os.ReadFile(autoBumperCfg)
	if err != nil {
		return fmt.Errorf("open autobumper config: %s", err)
	}

	var bumperClientOpt options
	err = yaml.Unmarshal(data, &bumperClientOpt)
	if err != nil {
		return fmt.Errorf("decode autobumper config: %s", err)
	}

	var opts bumper.Options
	err = yaml.Unmarshal(data, &opts)
	if err != nil {
		return fmt.Errorf("decode bumper options: %s", err)
	}

	ctx := context.Background()
	return bumper.Run(ctx, &opts, &client{o: &bumperClientOpt})
}
