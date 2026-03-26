package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

func TestDependencyAnalyzer_NoPackageJSON(t *testing.T) {
	a := &DependencyAnalyzer{}
	cfg := config.DefaultConfig()
	dir := t.TempDir()

	result, err := a.Analyze([]*parser.ParsedFile{}, cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 100 {
		t.Errorf("expected score=100 when no package.json, got %d", result.Score)
	}
}

func TestDependencyAnalyzer_CleanDeps(t *testing.T) {
	a := &DependencyAnalyzer{}
	cfg := config.DefaultConfig()
	dir := t.TempDir()

	pkg := `{
		"dependencies": {"express": "^4.18.0", "lodash": "^4.17.21"},
		"devDependencies": {"jest": "^29.0.0"}
	}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644)
	// Create a lock file
	os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("{}"), 0o644)

	result, err := a.Analyze([]*parser.ParsedFile{}, cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 100 {
		t.Errorf("expected score=100 for clean deps, got %d", result.Score)
	}
}

func TestDependencyAnalyzer_UnpinnedVersions(t *testing.T) {
	a := &DependencyAnalyzer{}
	cfg := config.DefaultConfig()
	dir := t.TempDir()

	pkg := `{
		"dependencies": {"bad-pkg": "*", "also-bad": "latest", "ok-pkg": "^1.0.0"}
	}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644)

	result, err := a.Analyze([]*parser.ParsedFile{}, cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Score >= 100 {
		t.Errorf("expected score < 100 for unpinned versions, got %d", result.Score)
	}
}

func TestDependencyAnalyzer_NoLockFile(t *testing.T) {
	a := &DependencyAnalyzer{}
	cfg := config.DefaultConfig()
	dir := t.TempDir()

	pkg := `{"dependencies": {"express": "^4.18.0"}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0o644)

	result, err := a.Analyze([]*parser.ParsedFile{}, cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	// Should flag missing lock file
	if result.Score >= 100 {
		t.Errorf("expected score < 100 for missing lock file, got %d", result.Score)
	}
}
