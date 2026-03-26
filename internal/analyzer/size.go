package analyzer

import (
	"fmt"
	"math"
	"sort"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type SizeAnalyzer struct{}

func (a *SizeAnalyzer) Name() config.MetricName {
	return config.MetricSize
}

func (a *SizeAnalyzer) Analyze(files []*parser.ParsedFile, cfg *config.Config, cwd string) (*Result, error) {
	maxFileLines := cfg.Thresholds.MaxFileLines
	maxFunctionLines := cfg.Thresholds.MaxFunctionLines
	var details []Detail
	filesOverThreshold := 0
	functionsOverThreshold := 0
	totalFunctions := 0
	totalLines := 0

	for _, file := range files {
		totalLines += file.LineCount

		if file.LineCount > maxFileLines {
			filesOverThreshold++
			details = append(details, Detail{
				File:    file.RelativePath,
				Message: fmt.Sprintf("File has %d lines (threshold: %d)", file.LineCount, maxFileLines),
				Value:   float64(file.LineCount),
			})
		}

		for _, fn := range file.Functions {
			totalFunctions++
			if fn.LineCount > maxFunctionLines {
				functionsOverThreshold++
				details = append(details, Detail{
					File:    file.RelativePath,
					Message: fmt.Sprintf("%s() has %d lines (threshold: %d)", fn.Name, fn.LineCount, maxFunctionLines),
					Line:    fn.StartLine,
					Value:   float64(fn.LineCount),
				})
			}
		}
	}

	totalItems := len(files) + totalFunctions
	violations := filesOverThreshold + functionsOverThreshold
	score := 100
	if totalItems > 0 {
		score = int(math.Round(float64(totalItems-violations) / float64(totalItems) * 100))
	}

	sort.Slice(details, func(i, j int) bool {
		return details[i].Value > details[j].Value
	})

	return &Result{
		Metric:  config.MetricSize,
		Score:   score,
		Details: details,
		Summary: fmt.Sprintf("%d total lines across %d files. %d oversized files, %d oversized functions", totalLines, len(files), filesOverThreshold, functionsOverThreshold),
	}, nil
}
