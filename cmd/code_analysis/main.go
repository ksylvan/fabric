package main

import (
	"fmt"
	"os"

	"github.com/danielmiessler/fabric/cmd/code_analysis/internal"
	"github.com/spf13/cobra"
)

var (
	outputFile string
	targetDir  string
)

var rootCmd = &cobra.Command{
	Use:   "code_analysis",
	Short: "Analyze codebase and generate metrics report",
	Long: `A code analysis tool that scans the codebase, collects metrics,
and generates a markdown report with file statistics and code health indicators.`,
	RunE:         run,
	SilenceUsage: true,
}

func init() {
	rootCmd.Flags().StringVarP(&targetDir, "directory", "d", ".", "Directory to analyze (default: current directory)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for the report (default: stdout)")
}

func run(cmd *cobra.Command, args []string) error {
	analyzer := internal.NewAnalyzer(targetDir)
	report, err := analyzer.Analyze()
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(report), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Report written to %s\n", outputFile)
	} else {
		fmt.Print(report)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
