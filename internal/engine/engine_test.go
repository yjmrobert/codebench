package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codebench/codebench/internal/analyzer"
	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/scoring"
	"github.com/codebench/codebench/internal/storage"
)

func TestAnalyze_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	report, warnings, err := Analyze(AnalysisOptions{
		Path:        dir,
		SaveHistory: false,
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Score == nil {
		t.Fatal("expected non-nil score")
	}
	if report.FilesAnalyzed != 0 {
		t.Errorf("expected 0 files analyzed, got %d", report.FilesAnalyzed)
	}
	// Warnings may or may not exist depending on analyzers
	_ = warnings
}

func TestAnalyze_WithJSFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a simple JS file
	jsContent := `function add(a, b) {
  return a + b;
}

function complex(x) {
  if (x > 0) {
    if (x > 10) {
      for (let i = 0; i < x; i++) {
        if (i > 5) {
          return i;
        }
      }
    }
  }
  return 0;
}
`
	if err := os.WriteFile(filepath.Join(dir, "index.js"), []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	report, _, err := Analyze(AnalysisOptions{
		Path:        dir,
		SaveHistory: false,
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if report.FilesAnalyzed != 1 {
		t.Errorf("expected 1 file analyzed, got %d", report.FilesAnalyzed)
	}

	if report.Score.Overall < 0 || report.Score.Overall > 100 {
		t.Errorf("score %d out of range [0, 100]", report.Score.Overall)
	}

	if len(report.Violations) == 0 {
		t.Log("No violations found (expected some from complex function)")
	}
}

func TestAnalyze_WithCustomConfig(t *testing.T) {
	dir := t.TempDir()

	jsContent := `function simple(x) { return x + 1; }
`
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	report, _, err := Analyze(AnalysisOptions{
		Path:        dir,
		Config:      cfg,
		SaveHistory: false,
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}
}

func TestAnalyze_SpecificMetric(t *testing.T) {
	dir := t.TempDir()

	jsContent := `function foo() { return 1; }
`
	if err := os.WriteFile(filepath.Join(dir, "test.js"), []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	report, _, err := Analyze(AnalysisOptions{
		Path:        dir,
		Metrics:     []string{"debt"},
		SaveHistory: false,
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if len(report.Results) != 1 {
		t.Errorf("expected 1 result for single metric, got %d", len(report.Results))
	}
	if report.Results[0].Metric != config.MetricDebt {
		t.Errorf("expected debt metric, got %s", report.Results[0].Metric)
	}
}

func TestAnalyze_InvalidMetric(t *testing.T) {
	dir := t.TempDir()

	_, _, err := Analyze(AnalysisOptions{
		Path:        dir,
		Metrics:     []string{"nonexistent"},
		SaveHistory: false,
	})
	if err == nil {
		t.Fatal("expected error for unknown metric")
	}
}

func TestAnalyze_SavesHistory(t *testing.T) {
	dir := t.TempDir()

	jsContent := `function foo() { return 1; }
`
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	report, _, err := Analyze(AnalysisOptions{
		Path:        dir,
		SaveHistory: true,
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// History should have been saved (RunID may be 0 if git not available, but DB should exist)
	dbPath := filepath.Join(dir, ".codebench", "history.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("expected history.db to be created")
	}
	_ = report
}

func TestViolationsSortedByImpact(t *testing.T) {
	cfg := config.DefaultConfig()

	results := []*analyzer.Result{
		{
			Metric: config.MetricCyclomatic,
			Score:  60,
			Details: []analyzer.Detail{
				{File: "a.js", Message: "high complexity", Value: 20, Severity: "critical", Category: "reduce-complexity", Threshold: 10},
				{File: "b.js", Message: "moderate complexity", Value: 12, Severity: "warning", Category: "reduce-complexity", Threshold: 10},
			},
		},
		{
			Metric: config.MetricDebt,
			Score:  90,
			Details: []analyzer.Detail{
				{File: "c.js", Message: "TODO: fix this", Severity: "info", Category: "resolve-todo"},
			},
		},
	}

	violations := buildViolations(results, cfg)

	if len(violations) != 3 {
		t.Fatalf("expected 3 violations, got %d", len(violations))
	}

	// Cyclomatic has weight 15 and score 60 -> deficit 40 -> per violation 20 -> impact = 20 * 0.15 = 3.0
	// Debt has weight 10 and score 90 -> deficit 10 -> per violation 10 -> impact = 10 * 0.10 = 1.0
	// So cyclomatic violations should come first
	if violations[0].Metric != "cyclomatic" {
		t.Errorf("expected first violation to be cyclomatic, got %s", violations[0].Metric)
	}
	if violations[2].Metric != "debt" {
		t.Errorf("expected last violation to be debt, got %s", violations[2].Metric)
	}

	// Verify sorted by impact descending
	for i := 0; i < len(violations)-1; i++ {
		if violations[i].Impact < violations[i+1].Impact {
			t.Errorf("violations not sorted by impact: [%d].Impact=%f < [%d].Impact=%f",
				i, violations[i].Impact, i+1, violations[i+1].Impact)
		}
	}
}

func TestCompare(t *testing.T) {
	before := &Report{
		Score: &scoring.CompositeScore{
			Overall:  70,
			SubGrade: "C",
			Metrics: []scoring.MetricScore{
				{Metric: config.MetricCoverage, Score: 60},
				{Metric: config.MetricCyclomatic, Score: 80},
			},
		},
	}

	after := &Report{
		Score: &scoring.CompositeScore{
			Overall:  85,
			SubGrade: "B",
			Metrics: []scoring.MetricScore{
				{Metric: config.MetricCoverage, Score: 80},
				{Metric: config.MetricCyclomatic, Score: 90},
			},
		},
	}

	delta := Compare(before, after)

	if delta.ScoreBefore != 70 {
		t.Errorf("expected ScoreBefore=70, got %d", delta.ScoreBefore)
	}
	if delta.ScoreAfter != 85 {
		t.Errorf("expected ScoreAfter=85, got %d", delta.ScoreAfter)
	}
	if delta.ScoreDelta != 15 {
		t.Errorf("expected ScoreDelta=15, got %d", delta.ScoreDelta)
	}
	if delta.GradeBefore != "C" {
		t.Errorf("expected GradeBefore=C, got %s", delta.GradeBefore)
	}
	if delta.GradeAfter != "B" {
		t.Errorf("expected GradeAfter=B, got %s", delta.GradeAfter)
	}
	if len(delta.MetricDeltas) != 2 {
		t.Fatalf("expected 2 metric deltas, got %d", len(delta.MetricDeltas))
	}
	if delta.MetricDeltas[0].Delta != 20 {
		t.Errorf("expected coverage delta=20, got %d", delta.MetricDeltas[0].Delta)
	}
}

func TestCompareToHistory_NoHistory(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".codebench", "history.db")

	report := &Report{
		Score: &scoring.CompositeScore{Overall: 80, SubGrade: "B-"},
	}

	delta, err := CompareToHistory(report, dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if delta != nil {
		t.Errorf("expected nil delta for no history, got %+v", delta)
	}
}

func TestCompareToHistory_WithHistory(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".codebench", "history.db")

	// Save a previous run
	prevScore := &scoring.CompositeScore{
		Overall:  70,
		Grade:    config.GradeC,
		SubGrade: "C",
		Metrics: []scoring.MetricScore{
			{Metric: config.MetricCoverage, Score: 65, Grade: "D+", Weight: 25},
		},
	}
	gitInfo := storage.GitInfo{CommitHash: "abc123", Branch: "main"}
	_, err := storage.SaveRun(dbPath, prevScore, gitInfo)
	if err != nil {
		t.Fatalf("SaveRun failed: %v", err)
	}

	// Create a new report with improved scores
	report := &Report{
		RunID: 0, // Different from saved run
		Score: &scoring.CompositeScore{
			Overall:  85,
			SubGrade: "B",
			Metrics: []scoring.MetricScore{
				{Metric: config.MetricCoverage, Score: 80, Grade: "B-", Weight: 25},
			},
		},
	}

	delta, err := CompareToHistory(report, dbPath)
	if err != nil {
		t.Fatalf("CompareToHistory failed: %v", err)
	}
	if delta == nil {
		t.Fatal("expected non-nil delta")
	}
	if delta.ScoreBefore != 70 {
		t.Errorf("expected ScoreBefore=70, got %d", delta.ScoreBefore)
	}
	if delta.ScoreAfter != 85 {
		t.Errorf("expected ScoreAfter=85, got %d", delta.ScoreAfter)
	}
	if delta.ScoreDelta != 15 {
		t.Errorf("expected ScoreDelta=15, got %d", delta.ScoreDelta)
	}
}

func TestBuildViolations_Empty(t *testing.T) {
	cfg := config.DefaultConfig()
	violations := buildViolations([]*analyzer.Result{}, cfg)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for empty results, got %d", len(violations))
	}
}

func TestBuildViolations_SeverityPreserved(t *testing.T) {
	cfg := config.DefaultConfig()
	results := []*analyzer.Result{
		{
			Metric: config.MetricCyclomatic,
			Score:  50,
			Details: []analyzer.Detail{
				{File: "a.js", Message: "high", Severity: "critical", Category: "reduce-complexity", Threshold: 10, Value: 25},
				{File: "b.js", Message: "med", Severity: "warning", Category: "reduce-complexity", Threshold: 10, Value: 15},
			},
		},
	}

	violations := buildViolations(results, cfg)
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	// Both should have same impact (same metric, same deficit split)
	if violations[0].Severity != "critical" && violations[1].Severity != "critical" {
		t.Error("expected at least one critical violation")
	}
}

func TestAnalyze_ProjectName(t *testing.T) {
	dir := t.TempDir()

	report, _, err := Analyze(AnalysisOptions{
		Path:        dir,
		SaveHistory: false,
	})
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if report.ProjectName == "" {
		t.Error("expected non-empty project name")
	}
}
