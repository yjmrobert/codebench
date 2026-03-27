package reporter

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/codebench/codebench/internal/analyzer"
	"github.com/codebench/codebench/internal/scoring"
)

var metricLabels = map[string]string{
	"coverage":     "Coverage",
	"cyclomatic":   "Cyclomatic",
	"cognitive":    "Cognitive",
	"duplication":  "Duplication",
	"dependencies": "Dependencies",
	"size":         "Size/Structure",
	"debt":         "Tech Debt",
}

func gradeColor(grade string) lipgloss.Color {
	switch {
	case strings.HasPrefix(grade, "A"):
		return lipgloss.Color("2") // green
	case strings.HasPrefix(grade, "B"):
		return lipgloss.Color("6") // cyan
	case strings.HasPrefix(grade, "C"):
		return lipgloss.Color("3") // yellow
	case strings.HasPrefix(grade, "D"):
		return lipgloss.Color("1") // red
	default:
		return lipgloss.Color("9") // bright red
	}
}

func progressBar(score, width int) string {
	filled := score * width / 100
	empty := width - filled
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	return green.Render(strings.Repeat("█", filled)) + gray.Render(strings.Repeat("░", empty))
}

func RenderTerminal(composite *scoring.CompositeScore, results []*analyzer.Result, projectName string) string {
	var sb strings.Builder

	// Header style
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	scoreColor := gradeColor(string(composite.Grade))
	scoreStyle := lipgloss.NewStyle().Bold(true).Foreground(scoreColor)

	// Title box
	titleBlock := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(0, 2).
		Width(50)

	title := headerStyle.Render(fmt.Sprintf("CodeBench Report — %s", projectName)) + "\n" +
		fmt.Sprintf("Score: %s  (%s)",
			scoreStyle.Render(fmt.Sprintf("%d / 100", composite.Overall)),
			scoreStyle.Render(composite.SubGrade))

	sb.WriteString("\n")
	sb.WriteString(titleBlock.Render(title))
	sb.WriteString("\n\n")

	// Metrics table
	rows := make([][]string, 0, len(composite.Metrics))
	for _, m := range composite.Metrics {
		label := metricLabels[string(m.Metric)]
		if label == "" {
			label = string(m.Metric)
		}
		bar := progressBar(m.Score, 10)
		color := gradeColor(m.Grade)
		gradeStyle := lipgloss.NewStyle().Foreground(color)
		rows = append(rows, []string{
			label,
			bar,
			fmt.Sprintf("%d%%", m.Score),
			gradeStyle.Render(fmt.Sprintf("(%s)", m.Grade)),
		})
	}

	t := table.New().
		Headers("Metric", "Progress", "Score", "Grade").
		Rows(rows...).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("4"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})

	sb.WriteString(t.Render())
	sb.WriteString("\n")

	// Warnings
	if len(composite.Warnings) > 0 {
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		sb.WriteString("\n")
		for _, w := range composite.Warnings {
			if len(w) > 60 {
				w = w[:57] + "..."
			}
			sb.WriteString(warnStyle.Render("  ⚠  " + w))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	return sb.String()
}
