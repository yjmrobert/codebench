package reporter

import (
	"fmt"
	"strings"

	"github.com/codebench/codebench/internal/analyzer"
	"github.com/codebench/codebench/internal/scoring"
)

var markdownMetricLabels = map[string]string{
	"coverage":     "Coverage",
	"cyclomatic":   "Cyclomatic Complexity",
	"cognitive":    "Cognitive Complexity",
	"duplication":  "Code Duplication",
	"dependencies": "Dependency Health",
	"size":         "Code Size & Structure",
	"debt":         "Technical Debt",
}

func RenderMarkdown(composite *scoring.CompositeScore, results []*analyzer.Result, projectName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## CodeBench Report — %s\n\n", projectName))
	sb.WriteString(fmt.Sprintf("**Score: %d / 100 (%s)**\n\n", composite.Overall, composite.SubGrade))
	sb.WriteString("| Metric | Score | Grade |\n")
	sb.WriteString("|--------|-------|-------|\n")

	for _, m := range composite.Metrics {
		label := markdownMetricLabels[string(m.Metric)]
		if label == "" {
			label = string(m.Metric)
		}
		sb.WriteString(fmt.Sprintf("| %s | %d%% | %s |\n", label, m.Score, m.Grade))
	}

	if len(composite.Warnings) > 0 {
		sb.WriteString("\n### Warnings\n\n")
		for _, w := range composite.Warnings {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", w))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}
