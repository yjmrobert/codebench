package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codebench/codebench/internal/config"
)

func TestDiscoverFiles_FindsMatchingFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ts"), []byte("const x = 1;"), 0o644)
	os.WriteFile(filepath.Join(dir, "utils.js"), []byte("const y = 2;"), 0o644)
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Hello"), 0o644)

	cfg := config.DefaultConfig()

	files, err := DiscoverFiles(cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestDiscoverFiles_RespectsIgnorePatterns(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ts"), []byte("const x = 1;"), 0o644)

	testDir := filepath.Join(dir, "node_modules", "pkg")
	os.MkdirAll(testDir, 0o755)
	os.WriteFile(filepath.Join(testDir, "index.js"), []byte("const z = 3;"), 0o644)

	cfg := config.DefaultConfig()

	files, err := DiscoverFiles(cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file (node_modules should be ignored), got %d", len(files))
	}
}

func TestDiscoverFiles_NoMatchingLanguage(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "main.rb"), []byte("puts 'hello'"), 0o644)

	cfg := config.DefaultConfig()

	files, err := DiscoverFiles(cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for unsupported language, got %d", len(files))
	}
}

func TestDiscoverFiles_EmptyExtensionSet(t *testing.T) {
	dir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Languages = []string{"unknown_lang"}

	files, err := DiscoverFiles(cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if files != nil {
		t.Errorf("expected nil files for unknown language, got %v", files)
	}
}
