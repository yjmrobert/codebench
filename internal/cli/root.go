package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/codebench/codebench/internal/engine"
	"github.com/codebench/codebench/internal/reporter"
)

var (
	flagJSON      bool
	flagFormat    string
	flagMetric    string
	flagThreshold int
	flagConfig    string
	flagNoCache   bool
	flagCompare   bool
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "codebench [path]",
		Short:   "Analyze codebases and produce a unified health score",
		Version: "0.1.0",
		Args:    cobra.MaximumNArgs(1),
		RunE:    runAnalysis,
	}

	rootCmd.Flags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.Flags().StringVar(&flagFormat, "format", "terminal", "Output format: terminal | json | markdown")
	rootCmd.Flags().StringVar(&flagMetric, "metric", "", "Run a single metric")
	rootCmd.Flags().IntVar(&flagThreshold, "threshold", 0, "Fail if score is below threshold")
	rootCmd.Flags().StringVar(&flagConfig, "config", "", "Path to config file")
	rootCmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "Skip cached results")
	rootCmd.Flags().BoolVar(&flagCompare, "compare", false, "Compare to previous run and include delta in output")

	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewHistoryCmd())

	return rootCmd
}

func runAnalysis(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if len(args) > 0 {
		target := args[0]
		if filepath.IsAbs(target) {
			cwd = target
		} else {
			cwd = filepath.Join(cwd, target)
		}
	}

	if flagThreshold < 0 || flagThreshold > 100 {
		return fmt.Errorf("--threshold must be between 0 and 100, got %d", flagThreshold)
	}

	format := flagFormat
	if flagJSON {
		format = "json"
	}

	// Parse metrics flag
	var metrics []string
	if flagMetric != "" {
		metrics = []string{flagMetric}
	}

	// Run analysis via engine
	report, warnings, err := engine.Analyze(engine.AnalysisOptions{
		Path:        cwd,
		ConfigPath:  flagConfig,
		Metrics:     metrics,
		SaveHistory: true,
	})
	if err != nil {
		return err
	}

	// Print analyzer warnings
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s analyzer failed: %v\n", w.Analyzer, w.Err)
	}

	if report.FilesAnalyzed == 0 && format == "terminal" {
		fmt.Fprintln(os.Stderr, "No source files found. Check your language settings and ignore patterns.")
		os.Exit(2)
	}

	// Compute delta if requested
	if flagCompare {
		dbPath := filepath.Join(cwd, ".codebench", "history.db")
		delta, err := engine.CompareToHistory(report, dbPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to compare to previous run: %v\n", err)
		}
		report.Delta = delta
	}

	// Render output
	switch format {
	case "json":
		output, err := reporter.RenderJSONReport(report)
		if err != nil {
			return err
		}
		fmt.Println(output)
	case "markdown":
		fmt.Print(reporter.RenderMarkdown(report.Score, report.Results, report.ProjectName))
	default:
		fmt.Print(reporter.RenderTerminal(report.Score, report.Results, report.ProjectName))
	}

	// Check threshold
	if flagThreshold > 0 && report.Score.Overall < flagThreshold {
		if format == "terminal" {
			fmt.Fprintf(os.Stderr, "Score %d is below threshold %d\n", report.Score.Overall, flagThreshold)
		}
		os.Exit(1)
	}

	return nil
}
