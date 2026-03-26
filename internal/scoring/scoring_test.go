package scoring

import (
	"testing"

	"github.com/codebench/codebench/internal/analyzer"
	"github.com/codebench/codebench/internal/config"
)

func TestToLetterGrade(t *testing.T) {
	tests := []struct {
		score int
		want  config.LetterGrade
	}{
		{100, config.GradeA},
		{95, config.GradeA},
		{90, config.GradeA},
		{89, config.GradeB},
		{80, config.GradeB},
		{79, config.GradeC},
		{70, config.GradeC},
		{69, config.GradeD},
		{60, config.GradeD},
		{59, config.GradeF},
		{0, config.GradeF},
	}

	for _, tt := range tests {
		got := ToLetterGrade(tt.score)
		if got != tt.want {
			t.Errorf("ToLetterGrade(%d) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func TestToSubGrade(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{97, "A+"},
		{95, "A"},
		{90, "A-"},
		{87, "B+"},
		{85, "B"},
		{80, "B-"},
		{50, "F"},
	}

	for _, tt := range tests {
		got := ToSubGrade(tt.score)
		if got != tt.want {
			t.Errorf("ToSubGrade(%d) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func makeResult(metric config.MetricName, score int) *analyzer.Result {
	return &analyzer.Result{
		Metric:  metric,
		Score:   score,
		Summary: "test",
	}
}

func TestComputeComposite(t *testing.T) {
	cfg := config.DefaultConfig()

	t.Run("weighted average", func(t *testing.T) {
		results := []*analyzer.Result{
			makeResult(config.MetricCoverage, 80),
			makeResult(config.MetricCyclomatic, 90),
			makeResult(config.MetricCognitive, 85),
			makeResult(config.MetricDuplication, 95),
			makeResult(config.MetricDependencies, 70),
			makeResult(config.MetricSize, 88),
			makeResult(config.MetricDebt, 75),
		}

		composite := ComputeComposite(results, cfg)
		// (80*25 + 90*15 + 85*15 + 95*15 + 70*10 + 88*10 + 75*10) / 100 = 83.8 -> 84
		if composite.Overall != 84 {
			t.Errorf("Overall = %d, want 84", composite.Overall)
		}
		if composite.Grade != config.GradeB {
			t.Errorf("Grade = %q, want B", composite.Grade)
		}
	})

	t.Run("all 100", func(t *testing.T) {
		var results []*analyzer.Result
		for _, m := range []config.MetricName{
			config.MetricCoverage, config.MetricCyclomatic, config.MetricCognitive,
			config.MetricDuplication, config.MetricDependencies, config.MetricSize, config.MetricDebt,
		} {
			results = append(results, makeResult(m, 100))
		}
		composite := ComputeComposite(results, cfg)
		if composite.Overall != 100 {
			t.Errorf("Overall = %d, want 100", composite.Overall)
		}
	})

	t.Run("all 0", func(t *testing.T) {
		var results []*analyzer.Result
		for _, m := range []config.MetricName{
			config.MetricCoverage, config.MetricCyclomatic, config.MetricCognitive,
			config.MetricDuplication, config.MetricDependencies, config.MetricSize, config.MetricDebt,
		} {
			results = append(results, makeResult(m, 0))
		}
		composite := ComputeComposite(results, cfg)
		if composite.Overall != 0 {
			t.Errorf("Overall = %d, want 0", composite.Overall)
		}
		if composite.Grade != config.GradeF {
			t.Errorf("Grade = %q, want F", composite.Grade)
		}
	})
}
