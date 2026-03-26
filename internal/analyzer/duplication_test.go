package analyzer

import (
	"strings"
	"testing"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

func makeDuplicationFile(content string) *parser.ParsedFile {
	lines := strings.Split(content, "\n")
	return &parser.ParsedFile{
		Path:         "/test/file.ts",
		RelativePath: "file.ts",
		Content:      content,
		Lines:        lines,
		LineCount:    len(lines),
		Language:     "typescript",
	}
}

func TestDuplicationAnalyzer_NoDuplication(t *testing.T) {
	a := &DuplicationAnalyzer{}
	content := "const a = 1;\nconst b = 2;\nconst c = 3;\nconst d = 4;\nconst e = 5;\nconst f = 6;\nconst g = 7;\nconst h = 8;\n"
	file := makeDuplicationFile(content)
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 100 {
		t.Errorf("expected score=100 for unique code, got %d", result.Score)
	}
}

func TestDuplicationAnalyzer_WithDuplication(t *testing.T) {
	a := &DuplicationAnalyzer{}
	// Create a block of 6+ identical lines appearing twice
	block := "const x = 1;\nconst y = 2;\nconst z = 3;\nconst w = 4;\nconst v = 5;\nconst u = 6;\n"
	content := block + "// separator\n" + block
	file := makeDuplicationFile(content)
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	if result.Score >= 100 {
		t.Errorf("expected score < 100 for duplicated code, got %d", result.Score)
	}
	if len(result.Details) == 0 {
		t.Error("expected duplication details")
	}
}

func TestDuplicationAnalyzer_EmptyFiles(t *testing.T) {
	a := &DuplicationAnalyzer{}
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 100 {
		t.Errorf("expected score=100 for no files, got %d", result.Score)
	}
}

func TestNormalizeLine(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"  const x = 1;  ", "const x = 1;"},
		{"", ""},
		{"{", ""},
		{"  a   b   c  ", "a b c"},
	}

	for _, tt := range tests {
		got := normalizeLine(tt.input)
		if got != tt.want {
			t.Errorf("normalizeLine(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
