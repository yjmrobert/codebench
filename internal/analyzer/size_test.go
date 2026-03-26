package analyzer

import (
	"strings"
	"testing"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

func makeSizeFile(lineCount int, functions []parser.FunctionInfo) *parser.ParsedFile {
	lines := make([]string, lineCount)
	for i := range lines {
		lines[i] = "const x = 1;"
	}
	return &parser.ParsedFile{
		Path:         "/test/file.ts",
		RelativePath: "file.ts",
		Content:      strings.Join(lines, "\n"),
		Lines:        lines,
		LineCount:    lineCount,
		Language:     "typescript",
		Functions:    functions,
	}
}

func TestSizeAnalyzer_SmallFile(t *testing.T) {
	a := &SizeAnalyzer{}
	file := makeSizeFile(50, []parser.FunctionInfo{
		{Name: "foo", StartLine: 1, EndLine: 10, LineCount: 10},
	})
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 100 {
		t.Errorf("expected score=100, got %d", result.Score)
	}
}

func TestSizeAnalyzer_OversizedFile(t *testing.T) {
	a := &SizeAnalyzer{}
	file := makeSizeFile(500, nil)
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 0 {
		t.Errorf("expected score=0, got %d", result.Score)
	}
	if len(result.Details) != 1 {
		t.Errorf("expected 1 detail, got %d", len(result.Details))
	}
}

func TestSizeAnalyzer_OversizedFunction(t *testing.T) {
	a := &SizeAnalyzer{}
	file := makeSizeFile(100, []parser.FunctionInfo{
		{Name: "bigFunc", StartLine: 1, EndLine: 80, LineCount: 80},
	})
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	// 1 file ok + 1 function over = 1/2 = 50%
	if result.Score != 50 {
		t.Errorf("expected score=50, got %d", result.Score)
	}
}
