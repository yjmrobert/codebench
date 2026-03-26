package analyzer

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type CoverageAnalyzer struct{}

type fileCoverage struct {
	file       string
	linesFound int
	linesHit   int
}

func parseLcov(content string) []fileCoverage {
	var files []fileCoverage
	var current fileCoverage
	inRecord := false

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "SF:") {
			current = fileCoverage{file: trimmed[3:]}
			inRecord = true
		} else if strings.HasPrefix(trimmed, "LF:") && inRecord {
			current.linesFound, _ = strconv.Atoi(trimmed[3:])
		} else if strings.HasPrefix(trimmed, "LH:") && inRecord {
			current.linesHit, _ = strconv.Atoi(trimmed[3:])
		} else if trimmed == "end_of_record" && inRecord {
			if current.file != "" {
				files = append(files, current)
			}
			inRecord = false
		}
	}

	return files
}

func (a *CoverageAnalyzer) Name() config.MetricName {
	return config.MetricCoverage
}

func (a *CoverageAnalyzer) Analyze(files []*parser.ParsedFile, cfg *config.Config, cwd string) (*Result, error) {
	reportPath := filepath.Join(cwd, cfg.Coverage.ReportPath)

	data, err := os.ReadFile(reportPath)
	if err != nil {
		return &Result{
			Metric:  config.MetricCoverage,
			Score:   0,
			Details: []Detail{{File: reportPath, Message: fmt.Sprintf("Coverage report not found at %s. Run your test suite with coverage enabled.", cfg.Coverage.ReportPath)}},
			Summary: "No coverage report found",
		}, nil
	}

	fileCoverages := parseLcov(string(data))
	if len(fileCoverages) == 0 {
		return &Result{
			Metric:  config.MetricCoverage,
			Score:   0,
			Details: []Detail{{File: reportPath, Message: "Coverage report is empty"}},
			Summary: "0% coverage (empty report)",
		}, nil
	}

	totalLines := 0
	totalHit := 0
	for _, f := range fileCoverages {
		totalLines += f.linesFound
		totalHit += f.linesHit
	}

	percentage := 0.0
	if totalLines > 0 {
		percentage = float64(totalHit) / float64(totalLines) * 100
	}

	var details []Detail
	for _, f := range fileCoverages {
		if f.linesFound > 0 {
			pct := float64(f.linesHit) / float64(f.linesFound) * 100
			details = append(details, Detail{
				File:    f.file,
				Message: fmt.Sprintf("%.1f%% coverage (%d/%d lines)", pct, f.linesHit, f.linesFound),
				Value:   pct,
			})
		}
	}

	sort.Slice(details, func(i, j int) bool {
		return details[i].Value < details[j].Value
	})

	return &Result{
		Metric:  config.MetricCoverage,
		Score:   int(math.Round(percentage)),
		Details: details,
		Summary: fmt.Sprintf("%.1f%% line coverage across %d files", percentage, len(fileCoverages)),
	}, nil
}
