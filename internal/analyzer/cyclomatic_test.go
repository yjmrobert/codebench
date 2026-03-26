package analyzer

import (
	"testing"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

func makeCyclomaticFile(functions []parser.FunctionInfo) *parser.ParsedFile {
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

func TestCyclomaticAnalyzer_SimpleFunction(t *testing.T) {
	a := &CyclomaticAnalyzer{}
	file := makeCyclomaticFile([]parser.FunctionInfo{
		{Name: "add", Body: "function add(a, b) { return a + b; }"},
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

func TestCyclomaticAnalyzer_ComplexFunction(t *testing.T) {
	a := &CyclomaticAnalyzer{}
	body := `function complex(x) {
		if (x > 0) {
			if (x > 10) {
				for (let i = 0; i < x; i++) {
					while (true) {
						if (i > 5 && x > 20) {
							break;
						}
					}
				}
			} else if (x > 5) {
				switch (x) {
					case 6: return "six";
					case 7: return "seven";
					case 8: return "eight";
				}
			}
		}
	}`
	file := makeCyclomaticFile([]parser.FunctionInfo{
		{Name: "complex", Body: body},
	})
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	// This function should exceed the default threshold of 10
	if result.Score >= 100 {
		t.Errorf("expected score < 100 for complex function, got %d", result.Score)
	}
	if len(result.Details) == 0 {
		t.Error("expected violation details for complex function")
	}
}

func TestCyclomaticAnalyzer_NoFunctions(t *testing.T) {
	a := &CyclomaticAnalyzer{}
	file := makeCyclomaticFile(nil)
	cfg := config.DefaultConfig()

	result, err := a.Analyze([]*parser.ParsedFile{file}, cfg, "/test")
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 100 {
		t.Errorf("expected score=100 when no functions, got %d", result.Score)
	}
}

func TestComputeCyclomaticComplexity(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"empty", "", 1},
		{"single if", "if (x) { }", 2},
		{"if-else", "if (x) { } else { }", 2},
		{"logical and", "if (a && b) { }", 3},
		{"logical or", "if (a || b) { }", 3},
		{"ternary", "const x = a ? b : c;", 2},
		{"for loop", "for (let i = 0; i < n; i++) { }", 2},
		{"while loop", "while (true) { }", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeCyclomaticComplexity(tt.body)
			if got != tt.want {
				t.Errorf("computeCyclomaticComplexity(%q) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}
