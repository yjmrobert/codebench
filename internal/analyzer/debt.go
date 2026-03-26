package analyzer

import (
	"fmt"
	"math"
	"regexp"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type DebtAnalyzer struct{}

var debtMarkerPattern = regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|XXX)\b`)

func (a *DebtAnalyzer) Name() config.MetricName {
	return config.MetricDebt
}

func (a *DebtAnalyzer) Analyze(files []*parser.ParsedFile, cfg *config.Config, cwd string) (*Result, error) {
	var details []Detail
	totalMarkers := 0
	totalLines := 0

	for _, file := range files {
		totalLines += file.LineCount

		for i, line := range file.Lines {
			matches := debtMarkerPattern.FindAllString(line, -1)
			if len(matches) > 0 {
				totalMarkers += len(matches)
				for _, match := range matches {
					trimmed := line
					if len(trimmed) > 100 {
						trimmed = trimmed[:100]
					}
					details = append(details, Detail{
						File:    file.RelativePath,
						Message: fmt.Sprintf("%s: %s", match, trimmed),
						Line:    i + 1,
					})
				}
			}
		}
	}

	density := 0.0
	if totalLines > 0 {
		density = float64(totalMarkers) / float64(totalLines) * 1000
	}

	maxDensity := cfg.Thresholds.MaxDebtDensity
	score := int(math.Round(math.Max(0, math.Min(100, 100-(density/maxDensity)*100))))

	// Cap details
	if len(details) > 50 {
		details = details[:50]
	}

	return &Result{
		Metric:  config.MetricDebt,
		Score:   score,
		Details: details,
		Summary: fmt.Sprintf("%d debt markers found (density: %.1f per 1K LOC)", totalMarkers, density),
	}, nil
}
