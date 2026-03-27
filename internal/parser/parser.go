package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type FunctionInfo struct {
	Name      string
	StartLine int
	EndLine   int
	LineCount int
	Body      string
}

type ParsedFile struct {
	Path         string
	RelativePath string
	Content      string
	Lines        []string
	LineCount    int
	Language     string
	Functions    []FunctionInfo
}

var extToLanguage = map[string]string{
	".js": "javascript", ".jsx": "javascript", ".mjs": "javascript", ".cjs": "javascript",
	".ts": "typescript", ".tsx": "typescript", ".mts": "typescript", ".cts": "typescript",
	".py": "python", ".go": "go", ".rs": "rust",
}

var reservedWords = map[string]bool{
	"if": true, "for": true, "while": true, "switch": true, "catch": true, "else": true,
}

var jsFuncPatterns = []*regexp.Regexp{
	// function declarations: function name(
	regexp.MustCompile(`(?:export\s+(?:default\s+)?)?(?:async\s+)?function\s+(\w+)\s*\(`),
	// arrow functions: const name = (...) => {
	regexp.MustCompile(`(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?(?:\([^)]*\)|[^=])\s*=>\s*\{`),
	// function expressions: const name = function(
	regexp.MustCompile(`(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?function\s*\(`),
	// class methods: name(...) {
	regexp.MustCompile(`(?m)^\s+(?:async\s+)?(?:static\s+)?(\w+)\s*\([^)]*\)\s*\{`),
}

var goFuncPatterns = []*regexp.Regexp{
	// func name(
	regexp.MustCompile(`(?m)^func\s+(\w+)\s*\(`),
	// method: func (receiver) name(
	regexp.MustCompile(`(?m)^func\s+\([^)]+\)\s+(\w+)\s*\(`),
}

var pythonFuncPattern = regexp.MustCompile(`(?m)^\s*def\s+(\w+)\s*\(`)

var rustFuncPatterns = []*regexp.Regexp{
	// fn name(
	regexp.MustCompile(`(?m)(?:pub\s+)?(?:async\s+)?fn\s+(\w+)\s*[<(]`),
}

func extractBraceFunctions(content string, lines []string, patterns []*regexp.Regexp) []FunctionInfo {
	var functions []FunctionInfo
	seenLines := make(map[int]bool)

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatchIndex(content, -1)
		for _, match := range matches {
			name := content[match[2]:match[3]]
			if reservedWords[name] {
				continue
			}

			startOffset := match[0]
			startLine := strings.Count(content[:startOffset], "\n") + 1

			if seenLines[startLine] {
				continue
			}

			// Find opening brace
			braceStart := strings.Index(content[match[0]:], "{")
			if braceStart == -1 {
				continue
			}
			braceStart += match[0]

			// Brace-match to find end
			depth := 0
			endOffset := braceStart
			for i := braceStart; i < len(content); i++ {
				if content[i] == '{' {
					depth++
				} else if content[i] == '}' {
					depth--
					if depth == 0 {
						endOffset = i
						break
					}
				}
			}

			endLine := strings.Count(content[:endOffset], "\n") + 1
			lineCount := endLine - startLine + 1

			var bodyLines []string
			if startLine-1 < len(lines) && endLine <= len(lines) {
				bodyLines = lines[startLine-1 : endLine]
			}
			body := strings.Join(bodyLines, "\n")

			seenLines[startLine] = true
			functions = append(functions, FunctionInfo{
				Name:      name,
				StartLine: startLine,
				EndLine:   endLine,
				LineCount: lineCount,
				Body:      body,
			})
		}
	}

	sort.Slice(functions, func(i, j int) bool {
		return functions[i].StartLine < functions[j].StartLine
	})

	return functions
}

func extractPythonFunctions(content string, lines []string) []FunctionInfo {
	var functions []FunctionInfo
	matches := pythonFuncPattern.FindAllStringSubmatchIndex(content, -1)

	for _, match := range matches {
		name := content[match[2]:match[3]]
		startOffset := match[0]
		startLine := strings.Count(content[:startOffset], "\n") + 1

		// Determine indentation level of the def line
		defLine := lines[startLine-1]
		baseIndent := len(defLine) - len(strings.TrimLeft(defLine, " \t"))

		// Find end by looking for next line at same or lesser indentation
		endLine := startLine
		for i := startLine; i < len(lines); i++ {
			line := lines[i]
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			if indent <= baseIndent && i > startLine {
				break
			}
			endLine = i + 1
		}

		lineCount := endLine - startLine + 1
		var bodyLines []string
		if startLine-1 < len(lines) && endLine <= len(lines) {
			bodyLines = lines[startLine-1 : endLine]
		}
		body := strings.Join(bodyLines, "\n")

		functions = append(functions, FunctionInfo{
			Name:      name,
			StartLine: startLine,
			EndLine:   endLine,
			LineCount: lineCount,
			Body:      body,
		})
	}

	return functions
}

func extractFunctions(content string, lines []string, language string) []FunctionInfo {
	switch language {
	case "javascript", "typescript":
		return extractBraceFunctions(content, lines, jsFuncPatterns)
	case "go":
		return extractBraceFunctions(content, lines, goFuncPatterns)
	case "rust":
		return extractBraceFunctions(content, lines, rustFuncPatterns)
	case "python":
		return extractPythonFunctions(content, lines)
	default:
		return nil
	}
}

func ParseFile(filePath, cwd string) (*ParsedFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	ext := filepath.Ext(filePath)
	language := extToLanguage[ext]
	if language == "" {
		language = "unknown"
	}

	rel, err := filepath.Rel(cwd, filePath)
	if err != nil {
		rel = filePath
	}

	functions := extractFunctions(content, lines, language)

	return &ParsedFile{
		Path:         filePath,
		RelativePath: rel,
		Content:      content,
		Lines:        lines,
		LineCount:    len(lines),
		Language:     language,
		Functions:    functions,
	}, nil
}

func ParseFiles(filePaths []string, cwd string) ([]*ParsedFile, error) {
	var files []*ParsedFile
	for _, fp := range filePaths {
		f, err := ParseFile(fp, cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", fp, err)
			continue
		}
		files = append(files, f)
	}
	return files, nil
}
