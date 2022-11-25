package main

import (
	"bufio"
	"context"
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

var _ bumper.PRHandler = (*client)(nil)

type client struct {
	o *options
}

// Changes returns a slice of functions, each one does some stuff, and
// returns commit message for the changes
func (c *client) Changes() []func(context.Context) (string, error) {
	return []func(context.Context) (string, error){
		func(ctx context.Context) (string, error) {
			return "Bumping index.md", nil
		},
	}
}

// PRTitleBody returns the body of the PR, this function runs after each commit
func (c *client) PRTitleBody() (string, string, error) {
	return "Update index.md" + "\n", "", nil
}

// options is the options for autobumper operations.
type options struct {
	GitHubRepo      string   `yaml:"gitHubRepo"`
	FoldersToFilter []string `yaml:"foldersToFilter"`
	FilesToFilter   []string `yaml:"filesToFilter"`
}

func main() {
	f, err := os.Create("docs/index.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	o, pro, err := parseOptions()
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to parse options")
	}

	startPath, err := os.Getwd()
	filepath.Walk(startPath, func(path string, info os.FileInfo, e error) error {
		if filterByFileExtension(path) && filterByFolderName(path, o) && filterByFileName(path, o) {
			mdLine := getDescription(path, o)
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

	ctx := context.Background()
	logrus.SetLevel(logrus.DebugLevel)
	if err := bumper.Run(ctx, pro, &client{o: o}); err != nil {
		logrus.WithError(err).Fatalf("failed to run the bumper tool")
	}
}

func getPathFromRepositoryRoot(path string, o *options) string {
	return strings.Split(path, o.GitHubRepo)[1]
}

func parseOptions() (*options, *bumper.Options, error) {
	var config string
	var labelsOverride []string
	var skipPullRequest bool

	var o options
	flag.StringVar(&config, "config", "", "The path to the config file for the autobumber.")
	flag.StringSliceVar(&labelsOverride, "labels-override", nil, "Override labels to be added to PR.")
	flag.BoolVar(&skipPullRequest, "skip-pullrequest", false, "")
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

func filterByFileExtension(path string) bool {
	return strings.Contains(path, ".md")
}

func filterByFolderName(path string, o *options) bool {
	pathFromRoot := getPathFromRepositoryRoot(path, o)
	for _, folderName := range o.FoldersToFilter {
		if strings.Contains(pathFromRoot, folderName) {
			return false
		}
	}
	return true
}

func filterByFileName(path string, o *options) bool {
	pathFromRoot := getPathFromRepositoryRoot(path, o)
	for _, fileName := range o.FilesToFilter {
		if pathFromRoot == fileName {
			return false
		}
	}
	return true
}

func getDescription(path string, o *options) string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return getPathFromRepositoryRoot(path, o)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	pathFromRoot := getPathFromRepositoryRoot(path, o)
	var description = ""
	for fileScanner.Scan() {
		if len(description) == 0 && strings.Contains(fileScanner.Text(), "#") {
			description = "[" + strings.Replace(fileScanner.Text(), "# ", "", 1) + "](" + pathFromRoot + ") - "
		} else if len(description) > 0 && !strings.Contains(fileScanner.Text(), "#") && len(fileScanner.Text()) > 0 {
			description += fileScanner.Text() + "\n\n"
			break
		}
	}

	if len(description) > 0 {
		return description
	}
	return getPathFromRepositoryRoot(path, o)
}
