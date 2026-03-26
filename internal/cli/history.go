package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/codebench/codebench/internal/reporter"
	"github.com/codebench/codebench/internal/storage"
)

var flagLimit int

func NewHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show score trend over time",
		RunE:  runHistory,
	}

	cmd.Flags().IntVar(&flagLimit, "limit", 20, "Number of entries to show")

	return cmd
}

func runHistory(cmd *cobra.Command, args []string) error {
	if flagLimit < 1 || flagLimit > 1000 {
		return fmt.Errorf("--limit must be between 1 and 1000, got %d", flagLimit)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	dbPath := filepath.Join(cwd, ".codebench", "history.db")

	runs, err := storage.GetHistory(dbPath, flagLimit)
	if err != nil {
		return fmt.Errorf("failed to read history: %w", err)
	}

	fmt.Print(reporter.RenderHistory(runs))
	return nil
}
