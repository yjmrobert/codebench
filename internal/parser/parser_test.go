package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_JavaScript(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.js")
	content := `function add(a, b) {
  return a + b;
}

const multiply = (a, b) => {
  return a * b;
};
`
	os.WriteFile(file, []byte(content), 0o644)

	parsed, err := ParseFile(file, dir)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Language != "javascript" {
		t.Errorf("expected language=javascript, got %s", parsed.Language)
	}
	if parsed.RelativePath != "test.js" {
		t.Errorf("expected relativePath=test.js, got %s", parsed.RelativePath)
	}
	if len(parsed.Functions) < 1 {
		t.Errorf("expected at least 1 function, got %d", len(parsed.Functions))
	}
}

func TestParseFile_UnknownLanguage(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.xyz")
	os.WriteFile(file, []byte("hello"), 0o644)

	parsed, err := ParseFile(file, dir)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Language != "unknown" {
		t.Errorf("expected language=unknown, got %s", parsed.Language)
	}
	if len(parsed.Functions) != 0 {
		t.Errorf("expected 0 functions for unknown language, got %d", len(parsed.Functions))
	}
}

func TestParseFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "empty.ts")
	os.WriteFile(file, []byte(""), 0o644)

	parsed, err := ParseFile(file, dir)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.LineCount != 1 {
		t.Errorf("expected lineCount=1 for empty file, got %d", parsed.LineCount)
	}
}

func TestParseFile_NonExistent(t *testing.T) {
	_, err := ParseFile("/nonexistent/file.ts", "/nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseFiles_SkipsErrors(t *testing.T) {
	dir := t.TempDir()
	goodFile := filepath.Join(dir, "good.js")
	os.WriteFile(goodFile, []byte("const x = 1;"), 0o644)

	files, err := ParseFiles([]string{goodFile, "/nonexistent/bad.js"}, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 parsed file, got %d", len(files))
	}
}

func TestExtractFunctions(t *testing.T) {
	content := `function foo() {
  return 1;
}

async function bar(x) {
  return x + 1;
}
`
	lines := []string{
		"function foo() {",
		"  return 1;",
		"}",
		"",
		"async function bar(x) {",
		"  return x + 1;",
		"}",
	}

	functions := extractFunctions(content, lines)
	if len(functions) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(functions))
	}
	if functions[0].Name != "foo" {
		t.Errorf("expected first function name=foo, got %s", functions[0].Name)
	}
	if functions[1].Name != "bar" {
		t.Errorf("expected second function name=bar, got %s", functions[1].Name)
	}
}
