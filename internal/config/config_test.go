package config

import (
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
