package reporter

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/codebench/codebench/internal/storage"
)

func trendArrow(current, previous int, hasPrevious bool) string {
	if !hasPrevious {
		return " "
	}
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	if current > previous {
		return green.Render("▲")
	}
	if current < previous {
		return red.Render("▼")
	}
	return gray.Render("─")
}

func RenderHistory(runs []storage.RunRecord) string {
	if len(runs) == 0 {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		return style.Render("No history found. Run `codebench` to record your first score.") + "\n"
	}

	// Reverse to chronological order for display
	reversed := make([]storage.RunRecord, len(runs))
	for i, r := range runs {
		reversed[len(runs)-1-i] = r
	}

	rows := make([][]string, 0, len(reversed))
	for i, run := range reversed {
		t, parseErr := time.Parse(time.RFC3339, run.Timestamp)
		date := run.Timestamp
		if parseErr == nil {
			date = t.Format("2006-01-02")
		}
		commitShort := run.CommitHash
		if len(commitShort) > 7 {
			commitShort = commitShort[:7]
		}

		color := gradeColor(run.OverallGrade)
		scoreStyle := lipgloss.NewStyle().Foreground(color)

		hasPrevious := i > 0
		previous := 0
		if hasPrevious {
			previous = reversed[i-1].OverallScore
		}

		rows = append(rows, []string{
			date,
			run.Branch,
			commitShort,
			scoreStyle.Render(fmt.Sprintf("%d", run.OverallScore)),
			scoreStyle.Render(run.OverallGrade),
			trendArrow(run.OverallScore, previous, hasPrevious),
		})
	}

	t := table.New().
		Headers("Date", "Branch", "Commit", "Score", "Grade", "Trend").
		Rows(rows...).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("4"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})

	var sb strings.Builder
	header := lipgloss.NewStyle().Bold(true)
	sb.WriteString("\n")
	sb.WriteString(header.Render("CodeBench History"))
	sb.WriteString("\n\n")
	sb.WriteString(t.Render())
	sb.WriteString("\n\n")
	return sb.String()
}
