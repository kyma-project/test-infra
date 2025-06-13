package main

import (
	"fmt"

	"github.com/kyma-project/test-infra/pkg/controllerfilters"
	"github.com/kyma-project/test-infra/pkg/github"
	"github.com/kyma-project/test-infra/pkg/github/actions"
	"github.com/kyma-project/test-infra/pkg/logging"
	"github.com/kyma-project/test-infra/pkg/pathsfilter"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Options holds all command-line flag values.
type Options struct {
	FiltersFile      string
	Base             string
	Head             string
	RepoPath         string
	EventName        string
	TargetBranchName string
}

var (
	rootCmd *cobra.Command
	opts    Options
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		_ = fmt.Errorf("error executing command: %w", err)
	}
}

func init() {
	rootCmd = &cobra.Command{
		Use:   "pathsfilter",
		Short: "A tool to filter changed file paths and branches based on YAML definitions.",
		Long: `pathsfilter detects changed files between two git refs and filters them
			against file and branch patterns defined in a YAML file. It's designed for use in GitHub Actions
			to conditionally run workflow jobs based on both file changes and event context (event type, target branch).`,
		RunE: func(_ *cobra.Command, _ []string) error {
			log := logging.NewLogger()
			defer func(log *zap.SugaredLogger) {
				err := log.Sync()
				if err != nil {
					fmt.Printf("error syncing logger: %v\n", err)
				}
			}(log)

			log.Infow("Starting paths filter process")

			gitRepo, err := github.NewRepository(opts.RepoPath)
			if err != nil {
				return fmt.Errorf("failed to initialize git repository adapter: %w", err)
			}

			outputWriter := actions.NewOutputWriter(log)

			log.Infow("Loading filter definitions", "path", opts.FiltersFile)

			definitions, err := controllerfilters.Load(opts.FiltersFile)
			if err != nil {
				return fmt.Errorf("failed to load filter definitions: %w", err)
			}

			jobMatcher := pathsfilter.NewJobMatcher(definitions, log)
			filterService := pathsfilter.NewFilterService(jobMatcher, gitRepo, outputWriter, log)

			if err := filterService.Run(opts.EventName, opts.TargetBranchName, opts.Base, opts.Head); err != nil {
				return fmt.Errorf("application run failed: %w", err)
			}

			log.Infow("Paths filter process completed successfully")

			return nil
		},
	}

	rootCmd.Flags().StringVarP(&opts.FiltersFile, "filters-file", "f", ".github/controller-test-filters.yaml", "Path to the YAML file with filter definitions")
	rootCmd.Flags().StringVarP(&opts.Base, "base", "b", "main", "Base git ref for comparison")
	rootCmd.Flags().StringVarP(&opts.Head, "head", "H", "HEAD", "Head git ref for comparison")
	rootCmd.Flags().StringVarP(&opts.RepoPath, "repo-path", "p", ".", "Path to the local repository root directory")
	rootCmd.Flags().StringVarP(&opts.EventName, "event-name", "e", "", "The name of the GitHub event (e.g., 'push', 'pull_request_target')")
	rootCmd.Flags().StringVarP(&opts.TargetBranchName, "target-branch-name", "t", "", "The target branch of the event (e.g., 'main', 'develop')")
}
