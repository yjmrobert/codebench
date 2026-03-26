package scoring

import (
	"fmt"
	"math"

	"github.com/codebench/codebench/internal/analyzer"
	"github.com/codebench/codebench/internal/config"
)

type MetricScore struct {
	Metric  config.MetricName `json:"metric"`
	Score   int               `json:"score"`
	Grade   string            `json:"grade"`
	Weight  int               `json:"weight"`
	Summary string            `json:"summary"`
}

type CompositeScore struct {
	Overall  int               `json:"overall"`
	Grade    config.LetterGrade `json:"grade"`
	SubGrade string            `json:"sub_grade"`
	Metrics  []MetricScore     `json:"metrics"`
	Warnings []string          `json:"warnings"`
}

func ToLetterGrade(score int) config.LetterGrade {
	switch {
	case score >= 90:
		return config.GradeA
	case score >= 80:
		return config.GradeB
	case score >= 70:
		return config.GradeC
	case score >= 60:
		return config.GradeD
	default:
		return config.GradeF
	}
}

func ToSubGrade(score int) string {
	grade := ToLetterGrade(score)
	if grade == config.GradeF {
		return "F"
	}
	withinBand := score % 10
	switch {
	case withinBand >= 7:
		return fmt.Sprintf("%s+", grade)
	case withinBand >= 3:
		return string(grade)
	default:
		return fmt.Sprintf("%s-", grade)
	}
}

func ComputeComposite(results []*analyzer.Result, cfg *config.Config) *CompositeScore {
	totalWeight := 0
	for _, w := range cfg.Weights {
		totalWeight += w
	}

	weightedSum := 0.0
	var metrics []MetricScore
	var warnings []string

	for _, result := range results {
		weight := cfg.Weights[result.Metric]
		normalizedWeight := 0.0
		if totalWeight > 0 {
			normalizedWeight = float64(weight) / float64(totalWeight)
		}
		weightedSum += float64(result.Score) * normalizedWeight

		metrics = append(metrics, MetricScore{
			Metric:  result.Metric,
			Score:   result.Score,
			Grade:   ToSubGrade(result.Score),
			Weight:  weight,
			Summary: result.Summary,
		})

		// Collect warnings from low-scoring metrics
		if result.Score < 70 {
			for i, detail := range result.Details {
				if i >= 3 {
					break
				}
				warnings = append(warnings, detail.Message)
			}
		}
	}

	overall := int(math.Round(weightedSum))

	if len(warnings) > 10 {
		warnings = warnings[:10]
	}

	return &CompositeScore{
		Overall:  overall,
		Grade:    ToLetterGrade(overall),
		SubGrade: ToSubGrade(overall),
		Metrics:  metrics,
		Warnings: warnings,
	}
}
