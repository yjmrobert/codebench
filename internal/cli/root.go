package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/codebench/codebench/internal/analyzer"
	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
	"github.com/codebench/codebench/internal/reporter"
	"github.com/codebench/codebench/internal/scoring"
	"github.com/codebench/codebench/internal/storage"
)

var (
	flagJSON      bool
	flagFormat    string
	flagMetric    string
	flagThreshold int
	flagConfig    string
	flagNoCache   bool
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

	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewHistoryCmd())

	return rootCmd
}

func runAnalysis(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	if len(args) > 0 {
		cwd = args[0]
		if !filepath.IsAbs(cwd) {
			wd, _ := os.Getwd()
			cwd = filepath.Join(wd, cwd)
		}
	}

	format := flagFormat
	if flagJSON {
		format = "json"
	}

	cfg, err := config.Load(cwd, flagConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Discover and parse files
	filePaths, err := parser.DiscoverFiles(cfg, cwd)
	if err != nil {
		return fmt.Errorf("file discovery failed: %w", err)
	}

	if len(filePaths) == 0 && format == "terminal" {
		fmt.Fprintln(os.Stderr, "No source files found. Check your language settings and ignore patterns.")
		os.Exit(2)
	}

	files, err := parser.ParseFiles(filePaths, cwd)
	if err != nil {
		return fmt.Errorf("file parsing failed: %w", err)
	}

	// Run analyzers
	allAnalyzers := analyzer.All()
	var targetAnalyzers []analyzer.Analyzer
	if flagMetric != "" {
		for _, a := range allAnalyzers {
			if string(a.Name()) == flagMetric {
				targetAnalyzers = append(targetAnalyzers, a)
			}
		}
		if len(targetAnalyzers) == 0 {
			fmt.Fprintf(os.Stderr, "Unknown metric: %s\n", flagMetric)
			os.Exit(2)
		}
	} else {
		targetAnalyzers = allAnalyzers
	}

	var results []*analyzer.Result
	for _, a := range targetAnalyzers {
		result, err := a.Analyze(files, cfg, cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s analyzer failed: %v\n", a.Name(), err)
			continue
		}
		results = append(results, result)
	}

	// Compute composite score
	composite := scoring.ComputeComposite(results, cfg)

	// Save to history
	dbPath := filepath.Join(cwd, ".codebench", "history.db")
	gitInfo := storage.GetGitInfo(cwd)
	storage.SaveRun(dbPath, composite, gitInfo) // non-fatal

	// Render output
	projectName := filepath.Base(cwd)
	switch format {
	case "json":
		output, err := reporter.RenderJSON(composite, results)
		if err != nil {
			return err
		}
		fmt.Println(output)
	case "markdown":
		fmt.Print(reporter.RenderMarkdown(composite, results, projectName))
	default:
		fmt.Print(reporter.RenderTerminal(composite, results, projectName))
	}

	// Check threshold
	if flagThreshold > 0 && composite.Overall < flagThreshold {
		if format == "terminal" {
			fmt.Fprintf(os.Stderr, "Score %d is below threshold %d\n", composite.Overall, flagThreshold)
		}
		os.Exit(1)
	}

	return nil
}
