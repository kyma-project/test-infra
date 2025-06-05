package main

import (
	"fmt"

	"github.com/kyma-project/test-infra/pkg/action"
	"github.com/kyma-project/test-infra/pkg/configloader"
	"github.com/kyma-project/test-infra/pkg/filter"
	"github.com/kyma-project/test-infra/pkg/git"
	"github.com/kyma-project/test-infra/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var config = viper.New()

var rootCmd = &cobra.Command{
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
				log.Errorw("Failed to sync logger", "error", err)
			}
		}(log)

		log.Infow("Starting paths filter process")

		gitRepo, err := git.NewRepository(config.GetString("working-dir"))
		if err != nil {
			return fmt.Errorf("failed to initialize git repository client: %w", err)
		}

		filtersPath := config.GetString("filters-file")

		log.Infow("Loading filter definitions", "path", filtersPath)

		definitions, err := configloader.Load(filtersPath)
		if err != nil {
			return fmt.Errorf("failed to load filter definitions: %w", err)
		}

		baseRef := config.GetString("base")
		headRef := config.GetString("head")

		log.Infow("Fetching changed files", "base", baseRef, "head", headRef)

		changedFiles, err := gitRepo.GetChangedFiles(baseRef, headRef)
		if err != nil {
			return fmt.Errorf("failed to get changed files: %w", err)
		}

		log.Infow("Found changed files", "count", len(changedFiles))
		log.Infow("Applying filters...")

		filterProcessor := filter.NewProcessor(definitions, log)
		filterResult := filterProcessor.Process(changedFiles)

		log.Infow("Found matching filters", "count", len(filterResult.MatchedFilterKeys))
		if config.GetBool("set-output") {
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		_ = fmt.Errorf("error executing command")
	}
}

func init() {
	rootCmd.Flags().StringP("filters-file", "f", ".github/controller-filters.yaml", "Path to the YAML file with filter definitions")
	rootCmd.Flags().StringP("base", "b", "main", "Base git ref for comparison")
	rootCmd.Flags().StringP("head", "H", "HEAD", "Head git ref for comparison")
	rootCmd.Flags().StringP("working-dir", "w", ".", "Working directory containing the .git repository")
	rootCmd.Flags().Bool("debug", false, "Enable debug logging")
	rootCmd.Flags().Bool("set-output", false, "Enable setting outputs for GitHub Actions")

	if err := config.BindPFlags(rootCmd.Flags()); err != nil {
		_ = fmt.Errorf("error binding flags: %w", err)
	}
}
