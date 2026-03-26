package reporter

import (
	"encoding/json"
	"time"

	"github.com/codebench/codebench/internal/analyzer"
	"github.com/codebench/codebench/internal/scoring"
)

type jsonMetric struct {
	Name    string            `json:"name"`
	Score   int               `json:"score"`
	Grade   string            `json:"grade"`
	Weight  int               `json:"weight"`
	Summary string            `json:"summary"`
	Details []analyzer.Detail `json:"details"`
}

type jsonReport struct {
	Version   string       `json:"version"`
	Timestamp string       `json:"timestamp"`
	Score     int          `json:"score"`
	Grade     string       `json:"grade"`
	Metrics   []jsonMetric `json:"metrics"`
	Warnings  []string     `json:"warnings"`
}

func RenderJSON(composite *scoring.CompositeScore, results []*analyzer.Result) (string, error) {
	report := jsonReport{
		Version:   "0.1.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Score:     composite.Overall,
		Grade:     composite.SubGrade,
		Warnings:  composite.Warnings,
	}

	if report.Warnings == nil {
		report.Warnings = []string{}
	}

	for _, m := range composite.Metrics {
		var details []analyzer.Detail
		for _, r := range results {
			if r.Metric == m.Metric {
				details = r.Details
				break
			}
		}
		if details == nil {
			details = []analyzer.Detail{}
		}

		report.Metrics = append(report.Metrics, jsonMetric{
			Name:    string(m.Metric),
			Score:   m.Score,
			Grade:   m.Grade,
			Weight:  m.Weight,
			Summary: m.Summary,
			Details: details,
		})
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
