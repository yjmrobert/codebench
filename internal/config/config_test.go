package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.Languages) != 2 {
		t.Errorf("expected 2 languages, got %d", len(cfg.Languages))
	}

	if cfg.Thresholds.MaxCyclomatic != 10 {
		t.Errorf("expected MaxCyclomatic=10, got %d", cfg.Thresholds.MaxCyclomatic)
	}
}

func TestDefaultWeightsSum(t *testing.T) {
	cfg := DefaultConfig()
	total := 0
	for _, w := range cfg.Weights {
		total += w
	}
	if total != 100 {
		t.Errorf("expected weights sum=100, got %d", total)
	}
}

func TestLoadNonexistent(t *testing.T) {
	cfg, err := Load("/nonexistent/path", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defaults := DefaultConfig()
	if cfg.Thresholds.MaxCyclomatic != defaults.Thresholds.MaxCyclomatic {
		t.Error("expected default config when file doesn't exist")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configPath := dir + "/.codebench.yml"
	os.WriteFile(configPath, []byte("{{invalid yaml"), 0o644)

	_, err := Load(dir, "")
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := dir + "/.codebench.yml"
	content := `languages:
  - python
thresholds:
  max_cyclomatic: 20
`
	os.WriteFile(configPath, []byte(content), 0o644)

	cfg, err := Load(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Languages) != 1 || cfg.Languages[0] != "python" {
		t.Errorf("expected [python], got %v", cfg.Languages)
	}
	if cfg.Thresholds.MaxCyclomatic != 20 {
		t.Errorf("expected MaxCyclomatic=20, got %d", cfg.Thresholds.MaxCyclomatic)
	}
	// Weights should be filled in with defaults
	if cfg.Weights[MetricCoverage] != 25 {
		t.Errorf("expected coverage weight=25, got %d", cfg.Weights[MetricCoverage])
	}
}

func TestLoadNegativeWeight(t *testing.T) {
	dir := t.TempDir()
	configPath := dir + "/.codebench.yml"
	content := `weights:
  coverage: -5
`
	os.WriteFile(configPath, []byte(content), 0o644)

	_, err := Load(dir, "")
	if err == nil {
		t.Error("expected error for negative weight, got nil")
	}
}
