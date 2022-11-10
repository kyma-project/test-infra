package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/test-infra/development/markdown-index/bumper"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

const repositoryName = "test-infra"

var _ bumper.PRHandler = (*client)(nil)

type client struct {
	o        *options
	images   map[string]string
	versions map[string][]string
}

// Changes returns a slice of functions, each one does some stuff, and
// returns commit message for the changes
func (c *client) Changes() []func(context.Context) (string, error) {
	return []func(context.Context) (string, error){
		func(ctx context.Context) (string, error) {
			return fmt.Sprintf("Bumping index.md"), nil
		},
	}
}

// PRTitleBody returns the body of the PR, this function runs after each commit
func (c *client) PRTitleBody() (string, string, error) {
	return "Update index.md" + "\n", "", nil
}

// prefix is the information needed for each prefix being bumped.
type prefix struct {
	// Name of the tool being bumped
	Name string `yaml:"name"`
	// The image prefix that the autobumper should look for
	Prefix string `yaml:"prefix"`
	// File that is looked at to determine current upstream image when bumping to upstream. Required only if targetVersion is "upstream"
	RefConfigFile string `yaml:"refConfigFile"`
	// File that is looked at to determine current upstream staging image when bumping to upstream staging. Required only if targetVersion is "upstream-staging"
	StagingRefConfigFile string `yaml:"stagingRefConfigFile"`
	// The repo where the image source resides for the images with this prefix. Used to create the links to see comparisons between images in the PR summary.
	Repo string `yaml:"repo"`
	// Whether or not the format of the PR summary for this prefix should be summarised.
	Summarise bool `yaml:"summarise"`
	// Whether the prefix tags should be consistent after the bump
	ConsistentImages bool `yaml:"consistentImages"`
}

// options is the options for autobumper operations.
type options struct {
	// The URL where upstream image references are located. Only required if Target Version is "upstream" or "upstreamStaging". Use "https://raw.githubusercontent.com/{ORG}/{REPO}"
	// Images will be bumped based off images located at the address using this URL and the refConfigFile or stagingRefConigFile for each Prefix.
	UpstreamURLBase string `yaml:"upstreamURLBase"`
	// The config paths to be included in this bump, in which only .yaml files will be considered. By default all files are included.
	IncludedConfigPaths []string `yaml:"includedConfigPaths"`
	// The config paths to be excluded in this bump, in which only .yaml files will be considered.
	ExcludedConfigPaths []string `yaml:"excludedConfigPaths"`
	// The extra non-yaml file to be considered in this bump.
	ExtraFiles []string `yaml:"extraFiles"`
	// The target version to bump images version to, which can be one of latest, upstream, upstream-staging and vYYYYMMDD-deadbeef.
	TargetVersion string `yaml:"targetVersion"`
	// List of prefixes that the autobumped is looking for, and other information needed to bump them. Must have at least 1 prefix.
	Prefixes []prefix `yaml:"prefixes"`
	// The oncall address where we can get the JSON file that stores the current oncall information.
	OncallAddress string `json:"onCallAddress"`
	// The oncall group that is responsible for reviewing the change, i.e. "test-infra".
	OncallGroup string `json:"onCallGroup"`
	// Whether skip if no oncall is discovered
	SkipIfNoOncall bool `yaml:"skipIfNoOncall"`
	// SkipOncallAssignment skips assigning to oncall.
	// The OncallAddress and OncallGroup are required for auto-bumper to figure out whether there are active oncall,
	// which is used to avoid bumping when there is no active oncall.
	SkipOncallAssignment bool `yaml:"skipOncallAssignment"`
	// SelfAssign is used to comment `/assign` and `/cc` so that blunderbuss wouldn't assign
	// bump PR to someone else.
	SelfAssign bool `yaml:"selfAssign"`
	// ImageRegistryAuth determines a way the autobumper with authenticate when talking to image registry.
	// Allowed values:
	// * "" (empty) -- uses no auth token
	// * "google" -- uses Google's "Application Default Credentials" as defined on https://pkg.go.dev/golang.org/x/oauth2/google#hdr-Credentials.
	ImageRegistryAuth string `yaml:"imageRegistryAuth"`
}

func main() {
	f, err := os.Create("index.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	startPath, err := os.Getwd()
	fmt.Println(startPath)
	filepath.Walk(startPath, func(path string, info os.FileInfo, e error) error {
		pathFromRepositoryRoot := strings.Split(path, repositoryName)[1]
		if filterByFileExtension(path) && filterByFolderName(path) && filterByFileName(pathFromRepositoryRoot) {
			mdLine := getDescription(path) + "\n[" + pathFromRepositoryRoot + "](" + pathFromRepositoryRoot + ")\n\n"
			fmt.Println(mdLine)
			//write line to file
			_, err = f.WriteString(mdLine)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	//bumper transplant
	ctx := context.Background()
	logrus.SetLevel(logrus.DebugLevel)
	o, pro, err := parseOptions()
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to run the bumper tool")
	}

	if err := validateOptions(o); err != nil {
		logrus.WithError(err).Fatalf("Failed validating flags")
	}

	if err := bumper.Run(ctx, pro, &client{o: o}); err != nil {
		logrus.WithError(err).Fatalf("failed to run the bumper tool")
	}
}

func parseOptions() (*options, *bumper.Options, error) {
	var config string
	var labelsOverride []string
	var skipPullRequest bool

	var o options
	flag.StringVar(&config, "config", "", "The path to the config file for the autobumber.")
	flag.StringSliceVar(&labelsOverride, "labels-override", nil, "Override labels to be added to PR.")
	flag.BoolVar(&skipPullRequest, "skip-pullrequest", false, "")
	flag.BoolVar(&o.SkipIfNoOncall, "skip-if-no-oncall", false, "Don't run anything if no oncall is discovered")
	flag.Parse()

	var pro bumper.Options
	data, err := os.ReadFile(config)
	if err != nil {
		return nil, nil, fmt.Errorf("read %q: %w", config, err)
	}

	if err = yaml.Unmarshal(data, &o); err != nil {
		return nil, nil, fmt.Errorf("unmarshal %q: %w", config, err)
	}

	if err := yaml.Unmarshal(data, &pro); err != nil {
		return nil, nil, fmt.Errorf("unmarshal %q: %w", config, err)
	}

	if labelsOverride != nil {
		pro.Labels = labelsOverride
	}
	pro.SkipPullRequest = skipPullRequest
	return &o, &pro, nil
}

func validateOptions(o *options) error {
	if len(o.Prefixes) == 0 {
		return errors.New("must have at least one Prefix specified")
	}
	if len(o.IncludedConfigPaths) == 0 {
		return errors.New("includedConfigPaths is mandatory")
	}

	return nil
}

func filterByFileExtension(path string) bool {
	return strings.Contains(path, ".md")
}

func filterByFolderName(path string) bool {
	return !strings.Contains(path, ".github") && !strings.Contains(path, ".githooks")
}

func filterByFileName(path string) bool {
	return path != "/CODE_OF_CONDUCT.md" && path != "/CONTRIBUTING.md" && path != "/NOTICE.md" && path != "/README.md" && path != "/index.md"
}

func getDescription(path string) string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return "# " + strings.Split(path, repositoryName)[1]
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	fileScanner.Split(bufio.ScanLines)

	var description = ""
	for fileScanner.Scan() {
		if len(description) == 0 && strings.Contains(fileScanner.Text(), "#") {
			description = fileScanner.Text() + "\n"
		} else if len(description) > 0 && !strings.Contains(fileScanner.Text(), "#") && len(fileScanner.Text()) > 0 {
			description += fileScanner.Text() + "\n"
			break
		}
	}

	if len(description) > 0 {
		return description
	}
	return "# " + strings.Split(path, repositoryName)[1]
}
