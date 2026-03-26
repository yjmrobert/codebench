export type MetricName =
  | 'coverage'
  | 'cyclomatic'
  | 'cognitive'
  | 'duplication'
  | 'dependencies'
  | 'size'
  | 'debt';

export type LetterGrade = 'A' | 'B' | 'C' | 'D' | 'F';

export interface CodeBenchConfig {
  languages: string[];
  weights: Record<MetricName, number>;
  thresholds: {
    max_cyclomatic: number;
    max_cognitive: number;
    max_file_lines: number;
    max_function_lines: number;
    min_coverage: number;
    max_duplication_pct: number;
    max_debt_density: number;
  };
  ignore: string[];
  coverage: {
    report_path: string;
  };
  ci: {
    fail_below: number;
    compare_branch: string;
  };
}
