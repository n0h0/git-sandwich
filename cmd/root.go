package cmd

import (
	"fmt"
	"os"
	"regexp"

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
)

var rootCmd = &cobra.Command{
	Use:   "git-sandwich [paths...]",
	Short: "Validate that changes are within BEGIN/END sandwich blocks",
	Long: `git-sandwich verifies that all changes in a Git diff are within
designated BEGIN/END blocks. Changes outside these blocks are rejected.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&startMarker, "start", "", "BEGIN marker regex (required)")
	rootCmd.Flags().StringVar(&endMarker, "end", "", "END marker regex (required)")
	rootCmd.Flags().StringVar(&baseRef, "base", "origin/main", "base ref for comparison")
	rootCmd.Flags().StringVar(&headRef, "head", "HEAD", "head ref for comparison")
	rootCmd.Flags().BoolVar(&allowNesting, "allow-nesting", false, "allow nested blocks")
	rootCmd.Flags().BoolVar(&allowBoundaryWithOutside, "allow-boundary-with-outside", false, "allow boundary changes with outside changes")
	rootCmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	_ = rootCmd.MarkFlagRequired("start")
	_ = rootCmd.MarkFlagRequired("end")
}
