import type { CodeBenchConfig } from './schema.js';

export const DEFAULT_CONFIG: CodeBenchConfig = {
  languages: ['javascript', 'typescript'],
  weights: {
    coverage: 25,
    cyclomatic: 15,
    cognitive: 15,
    duplication: 15,
    dependencies: 10,
    size: 10,
    debt: 10,
  },
  thresholds: {
    max_cyclomatic: 10,
    max_cognitive: 15,
    max_file_lines: 300,
    max_function_lines: 50,
    min_coverage: 80,
    max_duplication_pct: 5,
    max_debt_density: 2.0,
  },
  ignore: [
    'node_modules/**',
    'dist/**',
    'vendor/**',
    'coverage/**',
    '**/*.test.*',
    '**/*.spec.*',
    '**/*.d.ts',
  ],
  coverage: {
    report_path: './coverage/lcov.info',
  },
  ci: {
    fail_below: 70,
    compare_branch: 'main',
  },
};
