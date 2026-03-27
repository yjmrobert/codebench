package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type MetricName string

const (
	MetricCoverage     MetricName = "coverage"
	MetricCyclomatic   MetricName = "cyclomatic"
	MetricCognitive    MetricName = "cognitive"
	MetricDuplication  MetricName = "duplication"
	MetricDependencies MetricName = "dependencies"
	MetricSize         MetricName = "size"
	MetricDebt         MetricName = "debt"
)

type LetterGrade string

const (
	GradeA LetterGrade = "A"
	GradeB LetterGrade = "B"
	GradeC LetterGrade = "C"
	GradeD LetterGrade = "D"
	GradeF LetterGrade = "F"
)

type Thresholds struct {
	MaxCyclomatic    int     `yaml:"max_cyclomatic"`
	MaxCognitive     int     `yaml:"max_cognitive"`
	MaxFileLines     int     `yaml:"max_file_lines"`
	MaxFunctionLines int     `yaml:"max_function_lines"`
	MinCoverage      int     `yaml:"min_coverage"`
	MaxDuplicationPct float64 `yaml:"max_duplication_pct"`
	MaxDebtDensity   float64 `yaml:"max_debt_density"`
}

type CoverageConfig struct {
	ReportPath string `yaml:"report_path"`
}

type CIConfig struct {
	FailBelow     int    `yaml:"fail_below"`
	CompareBranch string `yaml:"compare_branch"`
}

type Config struct {
	Languages  []string              `yaml:"languages"`
	Weights    map[MetricName]int    `yaml:"weights"`
	Thresholds Thresholds            `yaml:"thresholds"`
	Ignore     []string              `yaml:"ignore"`
	Coverage   CoverageConfig        `yaml:"coverage"`
	CI         CIConfig              `yaml:"ci"`
}

func DefaultConfig() *Config {
	return &Config{
		Languages: []string{"javascript", "typescript"},
		Weights: map[MetricName]int{
			MetricCoverage:     25,
			MetricCyclomatic:   15,
			MetricCognitive:    15,
			MetricDuplication:  15,
			MetricDependencies: 10,
			MetricSize:         10,
			MetricDebt:         10,
		},
		Thresholds: Thresholds{
			MaxCyclomatic:    10,
			MaxCognitive:     15,
			MaxFileLines:     300,
			MaxFunctionLines: 50,
			MinCoverage:      80,
			MaxDuplicationPct: 5,
			MaxDebtDensity:   2.0,
		},
		Ignore: []string{
			"node_modules/**",
			"dist/**",
			"vendor/**",
			"coverage/**",
			"**/*.test.*",
			"**/*.spec.*",
			"**/*.d.ts",
		},
		Coverage: CoverageConfig{
			ReportPath: "./coverage/lcov.info",
		},
		CI: CIConfig{
			FailBelow:     70,
			CompareBranch: "main",
		},
	}
}

func Load(cwd, configPath string) (*Config, error) {
	if configPath == "" {
		configPath = filepath.Join(cwd, ".codebench.yml")
	}

	cfg := DefaultConfig()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Ensure weights map has all keys
	defaults := DefaultConfig()
	for k, v := range defaults.Weights {
		if _, ok := cfg.Weights[k]; !ok {
			cfg.Weights[k] = v
		}
	}

	// Validate weights are non-negative
	for k, v := range cfg.Weights {
		if v < 0 {
			return nil, fmt.Errorf("weight for %q must be non-negative, got %d", k, v)
		}
	}

	return cfg, nil
}

const SampleConfig = `# CodeBench Configuration

languages:
  - javascript
  - typescript

weights:
  coverage: 25
  cyclomatic: 15
  cognitive: 15
  duplication: 15
  dependencies: 10
  size: 10
  debt: 10

thresholds:
  max_cyclomatic: 10
  max_cognitive: 15
  max_file_lines: 300
  max_function_lines: 50
  min_coverage: 80
  max_duplication_pct: 5
  max_debt_density: 2.0

ignore:
  - "node_modules/**"
  - "dist/**"
  - "vendor/**"
  - "coverage/**"
  - "**/*.test.*"
  - "**/*.spec.*"
  - "**/*.d.ts"

coverage:
  report_path: "./coverage/lcov.info"

ci:
  fail_below: 70
  compare_branch: "main"
`
