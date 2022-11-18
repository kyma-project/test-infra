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

const repositoryName = "test-infra"

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
type options struct{}

func main() {
	f, err := os.Create("index.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Println(os.Getwd())
	path, err := os.Executable()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	startPath, err := os.Getwd()
	fmt.Println(startPath)
	filepath.Walk(startPath, func(path string, info os.FileInfo, e error) error {
		pathFromRepositoryRoot := strings.Split(path, repositoryName)[1]
		if filterByFileExtension(path) && filterByFolderName(path) && filterByFileName(pathFromRepositoryRoot) {
			mdLine := getDescription(path) + "\n[" + pathFromRepositoryRoot + "](" + pathFromRepositoryRoot + ")\n\n"
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
