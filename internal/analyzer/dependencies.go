package analyzer

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/codebench/codebench/internal/config"
	"github.com/codebench/codebench/internal/parser"
)

type DependencyAnalyzer struct{}

type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func (a *DependencyAnalyzer) Name() config.MetricName {
	return config.MetricDependencies
}

func (a *DependencyAnalyzer) Analyze(files []*parser.ParsedFile, cfg *config.Config, cwd string) (*Result, error) {
	pkgPath := filepath.Join(cwd, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return &Result{
			Metric:  config.MetricDependencies,
			Score:   100,
			Details: []Detail{{File: "package.json", Message: "No package.json found — skipping dependency analysis"}},
			Summary: "No package.json found",
		}, nil
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return &Result{
			Metric:  config.MetricDependencies,
			Score:   100,
			Details: []Detail{{File: "package.json", Message: "Failed to parse package.json"}},
			Summary: "Invalid package.json",
		}, nil
	}

	if pkg.Dependencies == nil {
		pkg.Dependencies = make(map[string]string)
	}
	if pkg.DevDependencies == nil {
		pkg.DevDependencies = make(map[string]string)
	}

	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	var details []Detail
	issues := 0

	// Check for unpinned versions
	for name, version := range allDeps {
		if version == "*" || version == "latest" {
			issues++
			details = append(details, Detail{
				File:     "package.json",
				Message:  fmt.Sprintf("%s uses unpinned version %q", name, version),
				Severity: "warning",
				Category: "pin-version",
			})
		}
	}

	const maxHealthyDependencies = 50

	// Check total dependency count
	totalDeps := len(pkg.Dependencies)
	totalDevDeps := len(pkg.DevDependencies)
	if totalDeps > maxHealthyDependencies {
		issues++
		details = append(details, Detail{
			File:      "package.json",
			Message:   fmt.Sprintf("%d production dependencies (consider reducing)", totalDeps),
			Value:     float64(totalDeps),
			Severity:  "warning",
			Category:  "reduce-dependencies",
			Threshold: float64(maxHealthyDependencies),
		})
	}

	// Check for lock file
	hasLock := false
	for _, lockFile := range []string{"package-lock.json", "yarn.lock", "pnpm-lock.yaml"} {
		if _, err := os.Stat(filepath.Join(cwd, lockFile)); err == nil {
			hasLock = true
			break
		}
	}
	if !hasLock {
		issues++
		details = append(details, Detail{
			File:     "package.json",
			Message:  "No lock file found — dependency versions may be non-deterministic",
			Severity: "warning",
			Category: "add-lockfile",
		})
	}

	maxIssues := len(allDeps)
	if maxIssues < 1 {
		maxIssues = 1
	}
	score := int(math.Round(math.Max(0, 100-float64(issues)/float64(maxIssues)*100)))
	if score > 100 {
		score = 100
	}

	return &Result{
		Metric:  config.MetricDependencies,
		Score:   score,
		Details: details,
		Summary: fmt.Sprintf("%d total dependencies (%d prod, %d dev), %d issues found", len(allDeps), totalDeps, totalDevDeps, issues),
	}, nil
}
