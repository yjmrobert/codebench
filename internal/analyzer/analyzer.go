package analyzer

import (
	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type Detail struct {
	File    string  `json:"file"`
	Message string  `json:"message"`
	Line    int     `json:"line,omitempty"`
	Value   float64 `json:"value,omitempty"`
}

type Result struct {
	Metric  config.MetricName `json:"metric"`
	Score   int               `json:"score"`
	Details []Detail          `json:"details"`
	Summary string            `json:"summary"`
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
