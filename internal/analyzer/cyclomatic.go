package analyzer

import (
	"fmt"
	"math"
	"regexp"
	"sort"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type CyclomaticAnalyzer struct{}

var cyclomaticPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bif\s*\(`),
	regexp.MustCompile(`\belse\s+if\s*\(`),
	regexp.MustCompile(`\bfor\s*\(`),
	regexp.MustCompile(`\bwhile\s*\(`),
	regexp.MustCompile(`\bdo\s*\{`),
	regexp.MustCompile(`\bcase\s+`),
	regexp.MustCompile(`\bcatch\s*\(`),
	regexp.MustCompile(`&&`),
	regexp.MustCompile(`\|\|`),
	regexp.MustCompile(`\?\?`),
	regexp.MustCompile(`\?[^?:]*:`),
}

var elseIfPattern = regexp.MustCompile(`\belse\s+if\s*\(`)

func computeCyclomaticComplexity(body string) int {
	complexity := 1

	for _, pattern := range cyclomaticPatterns {
		matches := pattern.FindAllString(body, -1)
		complexity += len(matches)
	}

	// Subtract double-counted else-if
	elseIfMatches := elseIfPattern.FindAllString(body, -1)
	complexity -= len(elseIfMatches)

	return complexity
}

func (a *CyclomaticAnalyzer) Name() config.MetricName {
	return config.MetricCyclomatic
}

func (a *CyclomaticAnalyzer) Analyze(files []*parser.ParsedFile, cfg *config.Config, cwd string) (*Result, error) {
	threshold := cfg.Thresholds.MaxCyclomatic
	var details []Detail
	totalFunctions := 0
	withinThreshold := 0

	for _, file := range files {
		for _, fn := range file.Functions {
			complexity := computeCyclomaticComplexity(fn.Body)
			totalFunctions++

			if complexity <= threshold {
				withinThreshold++
			} else {
				details = append(details, Detail{
					File:    file.RelativePath,
					Message: fmt.Sprintf("%s() has cyclomatic complexity %d (threshold: %d)", fn.Name, complexity, threshold),
					Line:    fn.StartLine,
					Value:   float64(complexity),
				})
			}
		}
	}

	if totalFunctions == 0 {
		return &Result{
			Metric:  config.MetricCyclomatic,
			Score:   100,
			Summary: "No functions found to analyze",
		}, nil
	}

	score := int(math.Round(float64(withinThreshold) / float64(totalFunctions) * 100))
	violations := totalFunctions - withinThreshold

	sort.Slice(details, func(i, j int) bool {
		return details[i].Value > details[j].Value
	})

	return &Result{
		Metric:  config.MetricCyclomatic,
		Score:   score,
		Details: details,
		Summary: fmt.Sprintf("%d of %d functions exceed complexity threshold of %d", violations, totalFunctions, threshold),
	}, nil
}
