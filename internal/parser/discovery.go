package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/codebench/codebench/internal/config"
)

var languageExtensions = map[string][]string{
	"javascript": {".js", ".jsx", ".mjs", ".cjs"},
	"typescript": {".ts", ".tsx", ".mts", ".cts"},
	"python":     {".py"},
	"go":         {".go"},
	"rust":       {".rs"},
}

func DiscoverFiles(cfg *config.Config, cwd string) ([]string, error) {
	extSet := make(map[string]bool)
	for _, lang := range cfg.Languages {
		if exts, ok := languageExtensions[lang]; ok {
			for _, ext := range exts {
				extSet[ext] = true
			}
		}
	}

	if len(extSet) == 0 {
		return nil, nil
	}

	var files []string
	err := filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error accessing %s: %v\n", path, err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(cwd, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot resolve relative path for %s: %v\n", path, err)
			return nil
		}

		// Check ignore patterns
		for _, pattern := range cfg.Ignore {
			matched, matchErr := doublestar.Match(pattern, rel)
			if matchErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: invalid ignore pattern %q: %v\n", pattern, matchErr)
				continue
			}
			if matched {
				return nil
			}
		}

		// Check extension
		ext := filepath.Ext(path)
		if extSet[ext] {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
