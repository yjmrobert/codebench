package engine

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/codebench/codebench/internal/analyzer"
	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
	"github.com/codebench/codebench/internal/scoring"
	"github.com/codebench/codebench/internal/storage"
)

// AnalysisOptions controls what gets analyzed.
type AnalysisOptions struct {
	Path        string         // Target directory (required)
	Config      *config.Config // If nil, loaded from ConfigPath or defaults
	ConfigPath  string         // Custom config file path (empty = auto-detect)
	Metrics     []string       // Specific metrics to run (empty = all)
	SaveHistory bool           // Whether to save to SQLite history
}

// Violation is a single actionable issue, enriched from analyzer.Detail.
type Violation struct {
	Metric    string  `json:"metric"`
	Severity  string  `json:"severity"`         // "critical", "warning", "info"
	File      string  `json:"file"`
	Line      int     `json:"line,omitempty"`
	Message   string  `json:"message"`
	Value     float64 `json:"value,omitempty"`
	Threshold float64 `json:"threshold,omitempty"`
	Category  string  `json:"category,omitempty"`
	Impact    float64 `json:"estimated_impact"` // estimated composite score improvement if fixed
}

// MetricDelta represents the change for a single metric between two runs.
type MetricDelta struct {
	Metric string `json:"metric"`
	Before int    `json:"before"`
	After  int    `json:"after"`
	Delta  int    `json:"delta"`
}

// Delta represents the change between two analysis runs.
type Delta struct {
	ScoreBefore  int           `json:"score_before"`
	ScoreAfter   int           `json:"score_after"`
	ScoreDelta   int           `json:"score_delta"`
	GradeBefore  string        `json:"grade_before"`
	GradeAfter   string        `json:"grade_after"`
	MetricDeltas []MetricDelta `json:"metric_deltas"`
}

// Report is the complete analysis result.
type Report struct {
	Score         *scoring.CompositeScore `json:"score"`
	Results       []*analyzer.Result      `json:"results"`
	Violations    []Violation             `json:"violations"`
	Delta         *Delta                  `json:"delta,omitempty"`
	RunID         int64                   `json:"run_id,omitempty"`
	GitInfo       storage.GitInfo         `json:"git_info"`
	FilesAnalyzed int                     `json:"files_analyzed"`
	ProjectName   string                  `json:"project_name"`
}

// AnalyzerWarning captures a non-fatal analyzer failure.
type AnalyzerWarning struct {
	Analyzer string
	Err      error
}

// Analyze runs the full analysis pipeline and returns structured results.
func Analyze(opts AnalysisOptions) (*Report, []AnalyzerWarning, error) {
	cfg := opts.Config
	if cfg == nil {
		var err error
		cfg, err = config.Load(opts.Path, opts.ConfigPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Discover and parse files
	filePaths, err := parser.DiscoverFiles(cfg, opts.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("file discovery failed: %w", err)
	}

	files, err := parser.ParseFiles(filePaths, opts.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("file parsing failed: %w", err)
	}

	// Determine which analyzers to run
	allAnalyzers := analyzer.All()
	var targetAnalyzers []analyzer.Analyzer
	if len(opts.Metrics) > 0 {
		metricSet := make(map[string]bool)
		for _, m := range opts.Metrics {
			metricSet[m] = true
		}
		for _, a := range allAnalyzers {
			if metricSet[string(a.Name())] {
				targetAnalyzers = append(targetAnalyzers, a)
			}
		}
		if len(targetAnalyzers) == 0 {
			names := make([]string, 0, len(allAnalyzers))
			for _, a := range allAnalyzers {
				names = append(names, string(a.Name()))
			}
			return nil, nil, fmt.Errorf("unknown metric(s) %v; available: %s", opts.Metrics, strings.Join(names, ", "))
		}
	} else {
		targetAnalyzers = allAnalyzers
	}

	// Run analyzers in parallel
	type analyzerResult struct {
		result *analyzer.Result
		err    error
		name   string
	}

	resultsCh := make([]analyzerResult, len(targetAnalyzers))
	var wg sync.WaitGroup
	for i, a := range targetAnalyzers {
		wg.Add(1)
		go func(idx int, an analyzer.Analyzer) {
			defer wg.Done()
			r, err := an.Analyze(files, cfg, opts.Path)
			resultsCh[idx] = analyzerResult{result: r, err: err, name: string(an.Name())}
		}(i, a)
	}
	wg.Wait()

	var results []*analyzer.Result
	var warnings []AnalyzerWarning
	for _, ar := range resultsCh {
		if ar.err != nil {
			warnings = append(warnings, AnalyzerWarning{Analyzer: ar.name, Err: ar.err})
			continue
		}
		results = append(results, ar.result)
	}

	// Compute composite score
	composite := scoring.ComputeComposite(results, cfg)

	// Build prioritized violations
	violations := buildViolations(results, cfg)

	// Optionally save to history
	var runID int64
	gitInfo := storage.GitInfo{CommitHash: "unknown", Branch: "unknown"}
	if opts.SaveHistory {
		dbPath := filepath.Join(opts.Path, ".codebench", "history.db")
		gitInfo = storage.GetGitInfo(opts.Path)
		runID, _ = storage.SaveRun(dbPath, composite, gitInfo)
	}

	report := &Report{
		Score:         composite,
		Results:       results,
		Violations:    violations,
		RunID:         runID,
		GitInfo:       gitInfo,
		FilesAnalyzed: len(files),
		ProjectName:   filepath.Base(opts.Path),
	}

	return report, warnings, nil
}

// CompareToHistory compares the current report to the most recent historical run.
func CompareToHistory(report *Report, dbPath string) (*Delta, error) {
	prevRun, prevMetrics, err := storage.GetPreviousRun(dbPath)
	if err != nil {
		return nil, err
	}
	if prevRun == nil {
		return nil, nil
	}

	// Don't compare to ourselves
	if report.RunID > 0 && int64(prevRun.ID) == report.RunID {
		return nil, nil
	}

	delta := &Delta{
		ScoreBefore: prevRun.OverallScore,
		ScoreAfter:  report.Score.Overall,
		ScoreDelta:  report.Score.Overall - prevRun.OverallScore,
		GradeBefore: prevRun.OverallGrade,
		GradeAfter:  report.Score.SubGrade,
	}

	for _, m := range report.Score.Metrics {
		before := 0
		if prevMetrics != nil {
			before = prevMetrics[string(m.Metric)]
		}
		delta.MetricDeltas = append(delta.MetricDeltas, MetricDelta{
			Metric: string(m.Metric),
			Before: before,
			After:  m.Score,
			Delta:  m.Score - before,
		})
	}

	return delta, nil
}

// Compare computes the delta between two reports.
func Compare(before, after *Report) *Delta {
	delta := &Delta{
		ScoreBefore: before.Score.Overall,
		ScoreAfter:  after.Score.Overall,
		ScoreDelta:  after.Score.Overall - before.Score.Overall,
		GradeBefore: before.Score.SubGrade,
		GradeAfter:  after.Score.SubGrade,
	}

	beforeMetrics := make(map[string]int)
	for _, m := range before.Score.Metrics {
		beforeMetrics[string(m.Metric)] = m.Score
	}

	for _, m := range after.Score.Metrics {
		b := beforeMetrics[string(m.Metric)]
		delta.MetricDeltas = append(delta.MetricDeltas, MetricDelta{
			Metric: string(m.Metric),
			Before: b,
			After:  m.Score,
			Delta:  m.Score - b,
		})
	}

	return delta
}

// buildViolations flattens all analyzer details into a prioritized violation list.
func buildViolations(results []*analyzer.Result, cfg *config.Config) []Violation {
	totalWeight := 0
	for _, w := range cfg.Weights {
		totalWeight += w
	}
	if totalWeight == 0 {
		totalWeight = 1
	}

	var violations []Violation
	for _, result := range results {
		weight := cfg.Weights[result.Metric]
		normalizedWeight := float64(weight) / float64(totalWeight)

		for _, d := range result.Details {
			impact := estimateImpact(d, result, normalizedWeight)
			violations = append(violations, Violation{
				Metric:    string(result.Metric),
				Severity:  d.Severity,
				File:      d.File,
				Line:      d.Line,
				Message:   d.Message,
				Value:     d.Value,
				Threshold: d.Threshold,
				Category:  d.Category,
				Impact:    math.Round(impact*100) / 100,
			})
		}
	}

	// Sort by impact descending, then severity
	severityOrder := map[string]int{"critical": 0, "warning": 1, "info": 2, "": 3}
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Impact != violations[j].Impact {
			return violations[i].Impact > violations[j].Impact
		}
		return severityOrder[violations[i].Severity] < severityOrder[violations[j].Severity]
	})

	return violations
}

// estimateImpact estimates how much fixing this violation would improve the composite score.
func estimateImpact(d analyzer.Detail, result *analyzer.Result, normalizedWeight float64) float64 {
	if len(result.Details) == 0 {
		return 0
	}
	// Each violation is roughly 1/N of the metric's score deficit
	scoreDeficit := float64(100 - result.Score)
	perViolation := scoreDeficit / float64(len(result.Details))
	// Multiply by metric weight to get composite impact
	return perViolation * normalizedWeight
}
