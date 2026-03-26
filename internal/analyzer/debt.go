package analyzer

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type DebtAnalyzer struct{}

const (
	maxLinePreviewLength = 100
	maxDebtDetails       = 50
)

var debtMarkerPattern = regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|XXX)\b`)

// commentPrefixPattern checks if a line contains a comment before the marker.
var commentPrefixPattern = regexp.MustCompile(`(?:^|\s)(?://|/\*|#|--|%|;)\s*`)

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
			matches := debtMarkerPattern.FindAllStringIndex(line, -1)
			if len(matches) == 0 {
				continue
			}

			// Only count markers that appear in comments or at word boundaries
			// not inside string literals or regex patterns
			for _, loc := range matches {
				prefix := line[:loc[0]]
				marker := line[loc[0]:loc[1]]

				// Skip if preceded by a hyphen (e.g., "resolve-todo")
				if loc[0] > 0 && line[loc[0]-1] == '-' {
					continue
				}

				// Skip if the marker appears inside backticks, which
				// suggests a string/regex literal
				if strings.Count(prefix, "`")%2 == 1 {
					continue
				}

				// Check if there's a comment indicator before this marker
				hasComment := commentPrefixPattern.MatchString(prefix)
				// Also accept markers at the start of a line (after whitespace)
				isLineStart := strings.TrimSpace(prefix) == ""

				if !hasComment && !isLineStart {
					continue
				}

				totalMarkers++
				trimmed := line
				if len(trimmed) > maxLinePreviewLength {
					trimmed = trimmed[:maxLinePreviewLength]
				}
				details = append(details, Detail{
					File:      file.RelativePath,
					Message:   fmt.Sprintf("%s: %s", strings.ToUpper(marker), strings.TrimSpace(trimmed)),
					Line:      i + 1,
					Severity:  "info",
					Category:  "resolve-todo",
					Threshold: cfg.Thresholds.MaxDebtDensity,
				})
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
	if len(details) > maxDebtDetails {
		details = details[:maxDebtDetails]
	}

	return &Result{
		Metric:  config.MetricDebt,
		Score:   score,
		Details: details,
		Summary: fmt.Sprintf("%d debt markers found (density: %.1f per 1K LOC)", totalMarkers, density),
	}, nil
}
