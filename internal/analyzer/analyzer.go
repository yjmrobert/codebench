package analyzer

import (
	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type Detail struct {
	File      string  `json:"file"`
	Message   string  `json:"message"`
	Line      int     `json:"line,omitempty"`
	Value     float64 `json:"value,omitempty"`
	Severity  string  `json:"severity,omitempty"`  // "critical", "warning", "info"
	Category  string  `json:"category,omitempty"`  // e.g., "reduce-complexity", "add-tests"
	Threshold float64 `json:"threshold,omitempty"` // the target value for this metric
}

type Result struct {
	Metric  config.MetricName `json:"metric"`
	Score   int               `json:"score"`
	Details []Detail          `json:"details"`
	Summary string            `json:"summary"`
}

// ComputeSeverity returns "critical" if value > 2*threshold, "warning" if > threshold, "info" otherwise.
// Use for metrics where higher values are worse (complexity, size, etc.).
func ComputeSeverity(value, threshold float64) string {
	if threshold <= 0 {
		return "info"
	}
	if value > 2*threshold {
		return "critical"
	}
	if value > threshold {
		return "warning"
	}
	return "info"
}

// ComputeSeverityInverse returns severity for metrics where lower values are worse (coverage).
// "critical" if value < threshold/2, "warning" if < threshold, "info" otherwise.
func ComputeSeverityInverse(value, threshold float64) string {
	if threshold <= 0 {
		return "info"
	}
	if value < threshold/2 {
		return "critical"
	}
	if value < threshold {
		return "warning"
	}
	return "info"
}

type Analyzer interface {
	Name() config.MetricName
	Analyze(files []*parser.ParsedFile, cfg *config.Config, cwd string) (*Result, error)
}

func All() []Analyzer {
	return []Analyzer{
		&CoverageAnalyzer{},
		&CyclomaticAnalyzer{},
		&CognitiveAnalyzer{},
		&DuplicationAnalyzer{},
		&DependencyAnalyzer{},
		&SizeAnalyzer{},
		&DebtAnalyzer{},
	}
}
