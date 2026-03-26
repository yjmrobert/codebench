package analyzer

import (
	"fmt"
	"math"
	"strings"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type DuplicationAnalyzer struct{}

const minDuplicateLines = 6

type normalizedLine struct {
	text         string
	originalLine int
}

func normalizeLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) <= 1 {
		return ""
	}
	// Collapse whitespace
	fields := strings.Fields(trimmed)
	return strings.Join(fields, " ")
}

func (a *DuplicationAnalyzer) Name() config.MetricName {
	return config.MetricDuplication
}

func (a *DuplicationAnalyzer) Analyze(files []*parser.ParsedFile, cfg *config.Config, cwd string) (*Result, error) {
	windowSize := minDuplicateLines

	type fileLines struct {
		file  *parser.ParsedFile
		lines []normalizedLine
	}

	var allFileLines []fileLines
	for _, file := range files {
		var normalized []normalizedLine
		for i, line := range file.Lines {
			n := normalizeLine(line)
			if n != "" {
				normalized = append(normalized, normalizedLine{text: n, originalLine: i + 1})
			}
		}
		allFileLines = append(allFileLines, fileLines{file: file, lines: normalized})
	}

	// Hash all windows
	type location struct {
		file string
		line int
	}
	windowMap := make(map[string][]location)
	totalNormalizedLines := 0

	for _, fl := range allFileLines {
		totalNormalizedLines += len(fl.lines)
		for i := 0; i <= len(fl.lines)-windowSize; i++ {
			var parts []string
			for j := 0; j < windowSize; j++ {
				parts = append(parts, fl.lines[i+j].text)
			}
			window := strings.Join(parts, "\n")

			loc := location{file: fl.file.RelativePath, line: fl.lines[i].originalLine}
			windowMap[window] = append(windowMap[window], loc)
		}
	}

	// Count duplicated lines
	duplicatedLines := 0
	seenDuplicates := make(map[string]bool)
	var details []Detail

	for _, locations := range windowMap {
		if len(locations) <= 1 {
			continue
		}

		for _, loc := range locations {
			key := fmt.Sprintf("%s:%d", loc.file, loc.line)
			if !seenDuplicates[key] {
				seenDuplicates[key] = true
				duplicatedLines += windowSize
			}
		}

		if len(details) < 20 {
			first := locations[0]
			for _, other := range locations[1:] {
				if len(details) >= 20 {
					break
				}
				details = append(details, Detail{
					File:    first.file,
					Message: fmt.Sprintf("Lines %d-%d duplicated in %s:%d", first.line, first.line+windowSize-1, other.file, other.line),
					Line:    first.line,
				})
			}
		}
	}

	duplicationPct := 0.0
	if totalNormalizedLines > 0 {
		duplicationPct = float64(duplicatedLines) / float64(totalNormalizedLines) * 100
	}

	maxPct := cfg.Thresholds.MaxDuplicationPct
	score := int(math.Round(math.Max(0, math.Min(100, 100-(duplicationPct/maxPct)*100))))

	return &Result{
		Metric:  config.MetricDuplication,
		Score:   score,
		Details: details,
		Summary: fmt.Sprintf("%.1f%% code duplication detected (%d duplicated lines)", duplicationPct, duplicatedLines),
	}, nil
}
