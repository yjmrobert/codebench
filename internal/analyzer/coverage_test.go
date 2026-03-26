package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

func TestParseLcov(t *testing.T) {
	content := `SF:src/index.ts
LF:100
LH:80
end_of_record
SF:src/utils.ts
LF:50
LH:50
end_of_record
`
	files := parseLcov(content)
	if len(files) != 2 {
		t.Fatalf("expected 2 file records, got %d", len(files))
	}
	if files[0].linesFound != 100 || files[0].linesHit != 80 {
		t.Errorf("first file: expected 100/80, got %d/%d", files[0].linesFound, files[0].linesHit)
	}
	if files[1].linesFound != 50 || files[1].linesHit != 50 {
		t.Errorf("second file: expected 50/50, got %d/%d", files[1].linesFound, files[1].linesHit)
	}
}

func TestParseLcov_Empty(t *testing.T) {
	files := parseLcov("")
	if len(files) != 0 {
		t.Errorf("expected 0 files from empty input, got %d", len(files))
	}
}

func TestCoverageAnalyzer_WithReport(t *testing.T) {
	a := &CoverageAnalyzer{}
	cfg := config.DefaultConfig()

	dir := t.TempDir()
	coverageDir := filepath.Join(dir, "coverage")
	os.MkdirAll(coverageDir, 0o755)

	lcov := `SF:src/index.ts
LF:100
LH:80
end_of_record
`
	os.WriteFile(filepath.Join(coverageDir, "lcov.info"), []byte(lcov), 0o644)

	result, err := a.Analyze([]*parser.ParsedFile{}, cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 80 {
		t.Errorf("expected score=80, got %d", result.Score)
	}
}

func TestCoverageAnalyzer_MissingReport(t *testing.T) {
	a := &CoverageAnalyzer{}
	cfg := config.DefaultConfig()
	dir := t.TempDir()

	result, err := a.Analyze([]*parser.ParsedFile{}, cfg, dir)
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 0 {
		t.Errorf("expected score=0 for missing report, got %d", result.Score)
	}
}
