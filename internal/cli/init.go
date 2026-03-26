package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/codebench/codebench/internal/config"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate a .codebench.yml config file",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, ".codebench.yml")
	dirPath := filepath.Join(cwd, ".codebench")

	if _, err := os.Stat(configPath); err == nil {
		fmt.Println(".codebench.yml already exists — skipping")
	} else {
		if err := os.WriteFile(configPath, []byte(config.SampleConfig), 0o644); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		fmt.Println("Created .codebench.yml")
	}

	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	fmt.Println("Created .codebench/ directory for history")

	fmt.Println("\nCodeBench initialized! Run `codebench` to analyze your project.")
	return nil
}
