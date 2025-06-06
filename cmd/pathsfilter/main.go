package main

import (
	"fmt"

	"github.com/kyma-project/test-infra/pkg/action"
	"github.com/kyma-project/test-infra/pkg/configloader"
	"github.com/kyma-project/test-infra/pkg/filter"
	"github.com/kyma-project/test-infra/pkg/git"
	"github.com/kyma-project/test-infra/pkg/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Options holds all command-line flag values for the application.
type Options struct {
	FiltersFile      string
	Base             string
	Head             string
	WorkingDirectory string
	Debug            bool
	SetOutput        bool
}

var (
	rootCmd *cobra.Command
	opts    Options
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		_ = fmt.Errorf("error executing command")
	}
}

func init() {
	rootCmd = &cobra.Command{
		Use:   "pathsfilter",
		Short: "A tool to filter changed file paths based on YAML definitions.",
		Long: `pathsfilter detects changed files between two git refs and filters them
against glob patterns defined in a YAML file. It's designed for use in GitHub Actions
to conditionally run workflow jobs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logging.NewLogger()
			defer func(log *zap.SugaredLogger) {
				err := log.Sync()
				if err != nil {
					fmt.Printf("error syncing log: %v\n", err)
				}
			}(log)

			log.Infow("Starting paths filter process")

			gitRepo, err := git.NewRepository(opts.WorkingDirectory)
			if err != nil {
				return fmt.Errorf("failed to initialize git repository client: %w", err)
			}

			log.Infow("Loading filter definitions", "path", opts.FiltersFile)
			definitions, err := configloader.Load(opts.FiltersFile)
			if err != nil {
				return fmt.Errorf("failed to load filter definitions: %w", err)
			}

			log.Infow("Fetching changed files", "base", opts.Base, "head", opts.Head)
			changedFiles, err := gitRepo.GetChangedFiles(opts.Base, opts.Head)
			if err != nil {
				return fmt.Errorf("failed to get changed files: %w", err)
			}
			log.Infow("Found changed files", "count", len(changedFiles))

			log.Infow("Applying filters...")
			filterProcessor := filter.NewProcessor(definitions, log)
			filterResult := filterProcessor.Process(changedFiles)
			log.Infow("Found matching filters", "count", len(filterResult.MatchedFilterKeys))

			if opts.SetOutput {
				log.Infow("Setting outputs for GitHub Actions")
				outputWriter := action.NewOutputWriter(log)
				if err := outputWriter.Write(filterResult); err != nil {
					return fmt.Errorf("failed to set action outputs: %w", err)
				}
			}

			log.Infow("Paths filter process completed successfully")

			return nil
		},
	}

	rootCmd.Flags().StringVarP(&opts.FiltersFile, "filters-file", "f", ".github/controller-filters.yaml", "Path to the YAML file with filter definitions")
	rootCmd.Flags().StringVarP(&opts.Base, "base", "b", "main", "Base git ref for comparison")
	rootCmd.Flags().StringVarP(&opts.Head, "head", "H", "HEAD", "Head git ref for comparison")
	rootCmd.Flags().StringVarP(&opts.WorkingDirectory, "working-dir", "w", ".", "Working directory containing the .git repository")
	rootCmd.Flags().BoolVar(&opts.Debug, "debug", false, "Enable debug logging")
	rootCmd.Flags().BoolVar(&opts.SetOutput, "set-output", false, "Enable setting outputs for GitHub Actions")
}
