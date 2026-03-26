package analyzer

import (
	"testing"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

func makeTestFile(content string) *parser.ParsedFile {
	lines := splitLines(content)
	return &parser.ParsedFile{
		Path:         "/test/file.ts",
		RelativePath: "file.ts",
		Content:      content,
		Lines:        lines,
		LineCount:    len(lines),
		Language:     "typescript",
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func TestDebtAnalyzer_FindsMarkers(t *testing.T) {
	a := &DebtAnalyzer{}
	file := makeTestFile("// TODO: fix this later\nconst x = 1;\n// FIXME: broken\n// HACK: temp workaround\n")
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Details) != 3 {
		t.Errorf("expected 3 details, got %d", len(result.Details))
	}
	if result.Score >= 100 {
		t.Errorf("expected score < 100, got %d", result.Score)
	}
}

func TestDebtAnalyzer_CleanCode(t *testing.T) {
	a := &DebtAnalyzer{}
	file := makeTestFile("const x = 1;\nconst y = 2;\nexport function add(a, b) { return a + b; }\n")
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Details) != 0 {
		t.Errorf("expected 0 details, got %d", len(result.Details))
	}
	if result.Score != 100 {
		t.Errorf("expected score=100, got %d", result.Score)
	}
}
