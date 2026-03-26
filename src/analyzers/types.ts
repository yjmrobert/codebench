import type { MetricName, LetterGrade, CodeBenchConfig } from '../config/index.js';
import type { ParsedFile } from '../parsers/index.js';

export interface AnalyzerDetail {
  file: string;
  message: string;
  line?: number;
  value?: number;
}

export interface AnalyzerResult {
  metric: MetricName;
  score: number; // 0-100
  details: AnalyzerDetail[];
  summary: string;
}

export interface Analyzer {
  name: MetricName;
  analyze(
    files: ParsedFile[],
    config: CodeBenchConfig,
    cwd: string,
  ): Promise<AnalyzerResult>;
}
