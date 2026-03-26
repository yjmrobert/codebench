import { writeFileSync, existsSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';
import yaml from 'js-yaml';
import { DEFAULT_CONFIG } from '../../config/index.js';

const SAMPLE_CONFIG = `# CodeBench Configuration
# See https://github.com/codebench/codebench for full documentation

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
  - "coverage/**"
  - "**/*.test.*"
  - "**/*.spec.*"
  - "**/*.d.ts"

coverage:
  report_path: "./coverage/lcov.info"

ci:
  fail_below: 70
  compare_branch: "main"
`;

export async function runInit(cwd: string): Promise<number> {
  const configPath = join(cwd, '.codebench.yml');
  const dirPath = join(cwd, '.codebench');

  if (existsSync(configPath)) {
    console.log('.codebench.yml already exists — skipping');
  } else {
    writeFileSync(configPath, SAMPLE_CONFIG, 'utf-8');
    console.log('Created .codebench.yml');
  }

  if (!existsSync(dirPath)) {
    mkdirSync(dirPath, { recursive: true });
    console.log('Created .codebench/ directory for history');
  }

  console.log('\nCodeBench initialized! Run `codebench` to analyze your project.');
  return 0;
}
