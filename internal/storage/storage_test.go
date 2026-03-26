package storage

import (
	"path/filepath"
	"testing"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/scoring"
)

func TestSaveAndGetHistory(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".codebench", "history.db")

	score := &scoring.CompositeScore{
		Overall:  85,
		Grade:    config.GradeB,
		SubGrade: "B",
		Metrics: []scoring.MetricScore{
			{Metric: config.MetricCoverage, Score: 80, Grade: "B-", Weight: 25},
		},
	}

	gitInfo := GitInfo{CommitHash: "abc1234", Branch: "main"}

	runID, err := SaveRun(dbPath, score, gitInfo)
	if err != nil {
		t.Fatalf("SaveRun failed: %v", err)
	}
	if runID == 0 {
		t.Error("expected non-zero run ID")
	}

	runs, err := GetHistory(dbPath, 10)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].OverallScore != 85 {
		t.Errorf("expected score=85, got %d", runs[0].OverallScore)
	}
	if runs[0].Branch != "main" {
		t.Errorf("expected branch=main, got %s", runs[0].Branch)
	}
	if runs[0].CommitHash != "abc1234" {
		t.Errorf("expected commit=abc1234, got %s", runs[0].CommitHash)
	}
}

func TestGetHistory_NonExistentDB(t *testing.T) {
	runs, err := GetHistory("/nonexistent/path/history.db", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runs != nil {
		t.Errorf("expected nil runs for nonexistent db, got %v", runs)
	}
}

func TestSaveMultipleRuns(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".codebench", "history.db")

	for i := 0; i < 3; i++ {
		score := &scoring.CompositeScore{
			Overall:  70 + i*10,
			Grade:    config.GradeC,
			SubGrade: "C",
		}
		gitInfo := GitInfo{CommitHash: "abc", Branch: "main"}
		_, err := SaveRun(dbPath, score, gitInfo)
		if err != nil {
			t.Fatalf("SaveRun %d failed: %v", i, err)
		}
	}

	runs, err := GetHistory(dbPath, 2)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs with limit=2, got %d", len(runs))
	}
}

func TestGetGitInfo(t *testing.T) {
	// Test with a non-git directory - should return "unknown"
	dir := t.TempDir()
	info := GetGitInfo(dir)
	if info.CommitHash != "unknown" {
		t.Errorf("expected unknown commit for non-git dir, got %s", info.CommitHash)
	}
}
