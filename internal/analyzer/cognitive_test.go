package analyzer

import (
	"testing"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

func makeCognitiveFile(functions []parser.FunctionInfo) *parser.ParsedFile {
	return &parser.ParsedFile{
		Path:         "/test/file.ts",
		RelativePath: "file.ts",
		Content:      "",
		Lines:        []string{},
		LineCount:    0,
		Language:     "typescript",
		Functions:    functions,
	}
}

func TestCognitiveAnalyzer_SimpleFunction(t *testing.T) {
	a := &CognitiveAnalyzer{}
	file := makeCognitiveFile([]parser.FunctionInfo{
		{Name: "add", Body: "function add(a, b) {\n  return a + b;\n}"},
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

func TestCognitiveAnalyzer_NestedFunction(t *testing.T) {
	a := &CognitiveAnalyzer{}
	body := `function nested(x) {
  if (x > 0) {
    for (let i = 0; i < x; i++) {
      if (i > 5) {
        while (true) {
          if (i > 10 && x > 20) {
            for (let j = 0; j < i; j++) {
              if (j > 3 || j < 1) {
                switch (j) {
                  case 1: break;
                }
              }
            }
          }
        }
      }
    }
  }
}`
	file := makeCognitiveFile([]parser.FunctionInfo{
		{Name: "nested", Body: body},
	})
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	// Deeply nested code should have high cognitive complexity
	if result.Score >= 100 {
		t.Errorf("expected score < 100 for nested function, got %d", result.Score)
	}
}

func TestCognitiveAnalyzer_NoFunctions(t *testing.T) {
	a := &CognitiveAnalyzer{}
	file := makeCognitiveFile(nil)
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 100 {
		t.Errorf("expected score=100 when no functions, got %d", result.Score)
	}
}

func TestComputeCognitiveComplexity(t *testing.T) {
	tests := []struct {
		name string
		body string
		min  int
	}{
		{"empty", "", 0},
		{"single if", "if (x) {\n}\n", 1},
		{"nested if", "if (x) {\n  if (y) {\n  }\n}\n", 3},
		{"logical ops", "if (a && b || c) {\n}\n", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeCognitiveComplexity(tt.body)
			if got < tt.min {
				t.Errorf("computeCognitiveComplexity(%q) = %d, want >= %d", tt.name, got, tt.min)
			}
		})
	}
}
