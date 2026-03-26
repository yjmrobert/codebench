# CodeBench

A CLI tool that analyzes codebases and produces a unified health score. CodeBench evaluates code quality across multiple dimensions and provides actionable insights to improve your project.

## Features

- **Composite Health Score** - A single 0-100 score with letter grades (A-F)
- **7 Built-in Metrics** - Coverage, cyclomatic complexity, cognitive complexity, duplication, dependencies, size, and technical debt
- **Multiple Output Formats** - Terminal (styled), JSON, and Markdown
- **History Tracking** - SQLite-backed run history with delta comparisons
- **Configurable** - Custom weights, thresholds, and ignore patterns via `.codebench.yml`
- **CI-Friendly** - Threshold-based exit codes for gating pipelines
- **Concurrent Analysis** - Runs all analyzers in parallel

## Installation

Requires Go 1.24+.

```sh
# Build from source
make build

# Install to $GOPATH/bin
make install
```

## Usage

```sh
# Analyze current directory
codebench

# Analyze a specific path
codebench ./path/to/project

# JSON output
codebench --json
# or
codebench --format json

# Markdown output
codebench --format markdown

# Run a single metric
codebench --metric cyclomatic

# Fail if score is below a threshold (useful for CI)
codebench --threshold 70

# Compare to the previous run
codebench --compare

# Use a custom config file
codebench --config path/to/.codebench.yml

# Skip cached results
codebench --no-cache
```

## Configuration

Create a `.codebench.yml` in your project root:

```yaml
languages:
  - javascript
  - typescript

weights:
  coverage: 25
  cyclomatic: 15
  cognitive: 15
  duplication: 15
  dependencies: 10
  size: 10
  debt: 10

thresholds:
  max_cyclomatic: 10
  max_cognitive: 15
  max_file_lines: 300
  max_function_lines: 50
  min_coverage: 80
  max_duplication_pct: 5
  max_debt_density: 2.0

ignore:
  - "node_modules/**"
  - "dist/**"
  - "vendor/**"

coverage:
  report_path: "./coverage/lcov.info"

ci:
  fail_below: 70
  compare_branch: "main"
```

You can also generate a starter config with:

```sh
codebench init
```

## Metrics

| Metric | Description |
|---|---|
| **Coverage** | Test coverage percentage from lcov reports |
| **Cyclomatic** | Cyclomatic complexity per function |
| **Cognitive** | Cognitive complexity per function |
| **Duplication** | Percentage of duplicated code |
| **Dependencies** | Dependency fan-in/fan-out analysis |
| **Size** | File and function length |
| **Debt** | Technical debt density |

## Subcommands

- `codebench init` - Generate a `.codebench.yml` config file
- `codebench history` - View past analysis runs

## Project Structure

```
cmd/codebench/       # CLI entrypoint
internal/
  analyzer/          # Individual metric analyzers
  cli/               # Cobra command definitions
  config/            # YAML config loading and defaults
  engine/            # Analysis orchestration and reporting API
  parser/            # File discovery and parsing
  reporter/          # Output renderers (terminal, JSON, markdown)
  scoring/           # Composite score computation
  storage/           # SQLite history storage
```

## Development

```sh
# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean
```

## License

See [LICENSE](LICENSE) for details.
