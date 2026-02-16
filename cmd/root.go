package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/n0h0/git-sandwich/internal/config"
	"github.com/n0h0/git-sandwich/internal/output"
	"github.com/n0h0/git-sandwich/internal/sandwich"
	"github.com/spf13/cobra"
)

var (
	startMarker              string
	endMarker                string
	baseRef                  string
	headRef                  string
	allowNesting             bool
	allowBoundaryWithOutside bool
	jsonOutput               bool
	includePatterns          []string
	excludePatterns          []string
	configPath               string
)

var rootCmd = &cobra.Command{
	Use:   "git-sandwich [paths...]",
	Short: "Validate that changes are within BEGIN/END sandwich blocks",
	Long: `git-sandwich verifies that all changes in a Git diff are within
designated BEGIN/END blocks. Changes outside these blocks are rejected.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := mergeConfig(cmd); err != nil {
			return err
		}

		startRe, err := regexp.Compile(startMarker)
		if err != nil {
			return fmt.Errorf("invalid --start regex: %w", err)
		}
		endRe, err := regexp.Compile(endMarker)
		if err != nil {
			return fmt.Errorf("invalid --end regex: %w", err)
		}

		cfg := &sandwich.Config{
			StartMarkerRegex:         startRe,
			EndMarkerRegex:           endRe,
			BaseRef:                  baseRef,
			HeadRef:                  headRef,
			AllowNesting:             allowNesting,
			AllowBoundaryWithOutside: allowBoundaryWithOutside,
			Paths:                    args,
			IncludePatterns:          includePatterns,
			ExcludePatterns:          excludePatterns,
		}

		result, err := sandwich.Validate(cfg)
		if err != nil {
			return err
		}

		if jsonOutput {
			if err := output.FormatJSON(os.Stdout, result); err != nil {
				return err
			}
		} else {
			output.FormatText(os.Stdout, result)
		}

		if !result.Success {
			os.Exit(1)
		}
		return nil
	},
}

func mergeConfig(cmd *cobra.Command) error {
	var fileCfg *config.FileConfig

	configExplicit := cmd.Flags().Changed("config")

	if configExplicit {
		// --config was explicitly specified: file must exist
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fileCfg = cfg
	} else {
		// Default path: load if exists, skip otherwise
		if _, err := os.Stat(configPath); err == nil {
			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			fileCfg = cfg
		}
	}

	if fileCfg != nil {
		if !cmd.Flags().Changed("start") && fileCfg.Start != "" {
			startMarker = fileCfg.Start
		}
		if !cmd.Flags().Changed("end") && fileCfg.End != "" {
			endMarker = fileCfg.End
		}
		if !cmd.Flags().Changed("base") && fileCfg.Base != "" {
			baseRef = fileCfg.Base
		}
		if !cmd.Flags().Changed("head") && fileCfg.Head != "" {
			headRef = fileCfg.Head
		}
		if !cmd.Flags().Changed("allow-nesting") && fileCfg.AllowNesting {
			allowNesting = fileCfg.AllowNesting
		}
		if !cmd.Flags().Changed("allow-boundary-with-outside") && fileCfg.AllowBoundaryWithOutside {
			allowBoundaryWithOutside = fileCfg.AllowBoundaryWithOutside
		}
		if !cmd.Flags().Changed("json") && fileCfg.JSON {
			jsonOutput = fileCfg.JSON
		}
		if !cmd.Flags().Changed("include") && len(fileCfg.Include) > 0 {
			includePatterns = fileCfg.Include
		}
		if !cmd.Flags().Changed("exclude") && len(fileCfg.Exclude) > 0 {
			excludePatterns = fileCfg.Exclude
		}
	}

	if startMarker == "" {
		return fmt.Errorf(`required flag "start" not set`)
	}
	if endMarker == "" {
		return fmt.Errorf(`required flag "end" not set`)
	}

	return nil
}

func SetVersion(v, c, d string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", v, c, d)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&startMarker, "start", "", "BEGIN marker regex")
	rootCmd.Flags().StringVar(&endMarker, "end", "", "END marker regex")
	rootCmd.Flags().StringVar(&baseRef, "base", "origin/main", "base ref for comparison")
	rootCmd.Flags().StringVar(&headRef, "head", "HEAD", "head ref for comparison")
	rootCmd.Flags().BoolVar(&allowNesting, "allow-nesting", false, "allow nested blocks")
	rootCmd.Flags().BoolVar(&allowBoundaryWithOutside, "allow-boundary-with-outside", false, "allow boundary changes with outside changes")
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.Flags().StringArrayVar(&includePatterns, "include", nil, "glob pattern for files to include (repeatable)")
	rootCmd.Flags().StringArrayVar(&excludePatterns, "exclude", nil, "glob pattern for files to exclude (repeatable)")
	rootCmd.Flags().StringVar(&configPath, "config", ".git-sandwich.yml", "path to config file")
}
