package analyzer

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type CognitiveAnalyzer struct{}

var (
	cogIfPattern     = regexp.MustCompile(`\b(if|else if)\b`)
	cogElsePattern   = regexp.MustCompile(`\belse\s*\{`)
	cogLoopPattern   = regexp.MustCompile(`\b(for|while)\b`)
	cogDoPattern     = regexp.MustCompile(`\bdo\s*\{`)
	cogSwitchPattern = regexp.MustCompile(`\bswitch\b`)
	cogCatchPattern  = regexp.MustCompile(`\bcatch\b`)
	cogLogicalPattern = regexp.MustCompile(`&&|\|\||\?\?`)
	cogTernaryPattern = regexp.MustCompile(`\?[^?:]*:`)
	braceOpenPattern  = regexp.MustCompile(`\{`)
	braceClosePattern = regexp.MustCompile(`\}`)
)

func computeCognitiveComplexity(body string) int {
	complexity := 0
	nestingDepth := 0
	lines := strings.Split(body, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		opens := len(braceOpenPattern.FindAllString(trimmed, -1))
		closes := len(braceClosePattern.FindAllString(trimmed, -1))

		if cogIfPattern.MatchString(trimmed) {
			complexity += 1 + nestingDepth
		} else if cogElsePattern.MatchString(trimmed) {
			complexity += 1
		}

		if cogLoopPattern.MatchString(trimmed) {
			complexity += 1 + nestingDepth
		}

		if cogDoPattern.MatchString(trimmed) {
			complexity += 1 + nestingDepth
		}

		if cogSwitchPattern.MatchString(trimmed) {
			complexity += 1 + nestingDepth
		}

		if cogCatchPattern.MatchString(trimmed) {
			complexity += 1 + nestingDepth
		}

		logicalOps := len(cogLogicalPattern.FindAllString(trimmed, -1))
		complexity += logicalOps

		if cogTernaryPattern.MatchString(trimmed) {
			complexity += 1 + nestingDepth
		}

		nestingDepth += opens - closes
		if nestingDepth < 0 {
			nestingDepth = 0
		}
	}

	return complexity
}

func (a *CognitiveAnalyzer) Name() config.MetricName {
	return config.MetricCognitive
}

func (a *CognitiveAnalyzer) Analyze(files []*parser.ParsedFile, cfg *config.Config, cwd string) (*Result, error) {
	threshold := cfg.Thresholds.MaxCognitive
	var details []Detail
	totalFunctions := 0
	withinThreshold := 0

	for _, file := range files {
		for _, fn := range file.Functions {
			complexity := computeCognitiveComplexity(fn.Body)
			totalFunctions++

			if complexity <= threshold {
				withinThreshold++
			} else {
				details = append(details, Detail{
					File:    file.RelativePath,
					Message: fmt.Sprintf("%s() has cognitive complexity %d (threshold: %d)", fn.Name, complexity, threshold),
					Line:    fn.StartLine,
					Value:   float64(complexity),
					Severity:  ComputeSeverity(float64(complexity), float64(threshold)),
					Category:  "reduce-nesting",
					Threshold: float64(threshold),
				})
			}
		}
	}

	if totalFunctions == 0 {
		return &Result{
			Metric:  config.MetricCognitive,
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
		Metric:  config.MetricCognitive,
		Score:   score,
		Details: details,
		Summary: fmt.Sprintf("%d of %d functions exceed cognitive complexity threshold of %d", violations, totalFunctions, threshold),
	}, nil
}
