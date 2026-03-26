package reporter

import (
	"encoding/json"
	"time"

	"github.com/codebench/codebench/internal/analyzer"
	"github.com/codebench/codebench/internal/engine"
	"github.com/codebench/codebench/internal/scoring"
)

type jsonViolation struct {
	Metric    string  `json:"metric"`
	Severity  string  `json:"severity"`
	File      string  `json:"file"`
	Line      int     `json:"line,omitempty"`
	Message   string  `json:"message"`
	Value     float64 `json:"value,omitempty"`
	Threshold float64 `json:"threshold,omitempty"`
	Category  string  `json:"category,omitempty"`
	Impact    float64 `json:"estimated_impact"`
}

type jsonMetricDelta struct {
	Metric string `json:"metric"`
	Before int    `json:"before"`
	After  int    `json:"after"`
	Delta  int    `json:"delta"`
}

type jsonDelta struct {
	ScoreBefore  int               `json:"score_before"`
	ScoreAfter   int               `json:"score_after"`
	ScoreDelta   int               `json:"score_delta"`
	GradeBefore  string            `json:"grade_before"`
	GradeAfter   string            `json:"grade_after"`
	MetricDeltas []jsonMetricDelta `json:"metric_deltas"`
}

type jsonMetric struct {
	Name    string            `json:"name"`
	Score   int               `json:"score"`
	Grade   string            `json:"grade"`
	Weight  int               `json:"weight"`
	Summary string            `json:"summary"`
	Details []analyzer.Detail `json:"details"`
}

type jsonReport struct {
	Version    string          `json:"version"`
	Timestamp  string          `json:"timestamp"`
	Score      int             `json:"score"`
	Grade      string          `json:"grade"`
	Metrics    []jsonMetric    `json:"metrics"`
	Warnings   []string        `json:"warnings"`
	Violations []jsonViolation `json:"violations"`
	Delta      *jsonDelta      `json:"delta,omitempty"`
}

// RenderJSON renders a JSON report from composite score and results (legacy interface).
func RenderJSON(composite *scoring.CompositeScore, results []*analyzer.Result) (string, error) {
	report := buildJSONReport(composite, results, nil, nil)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RenderJSONReport renders the full JSON report from an engine.Report, including violations and delta.
func RenderJSONReport(r *engine.Report) (string, error) {
	var violations []jsonViolation
	for _, v := range r.Violations {
		violations = append(violations, jsonViolation{
			Metric:    v.Metric,
			Severity:  v.Severity,
			File:      v.File,
			Line:      v.Line,
			Message:   v.Message,
			Value:     v.Value,
			Threshold: v.Threshold,
			Category:  v.Category,
			Impact:    v.Impact,
		})
	}

	var delta *jsonDelta
	if r.Delta != nil {
		delta = &jsonDelta{
			ScoreBefore: r.Delta.ScoreBefore,
			ScoreAfter:  r.Delta.ScoreAfter,
			ScoreDelta:  r.Delta.ScoreDelta,
			GradeBefore: r.Delta.GradeBefore,
			GradeAfter:  r.Delta.GradeAfter,
		}
		for _, md := range r.Delta.MetricDeltas {
			delta.MetricDeltas = append(delta.MetricDeltas, jsonMetricDelta{
				Metric: md.Metric,
				Before: md.Before,
				After:  md.After,
				Delta:  md.Delta,
			})
		}
	}

	report := buildJSONReport(r.Score, r.Results, violations, delta)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildJSONReport(composite *scoring.CompositeScore, results []*analyzer.Result, violations []jsonViolation, delta *jsonDelta) jsonReport {
	report := jsonReport{
		Version:   "0.2.0",
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

	if violations != nil {
		report.Violations = violations
	} else {
		report.Violations = []jsonViolation{}
	}

	report.Delta = delta

	return report
}
