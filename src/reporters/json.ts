import type { CompositeScore } from '../scoring/index.js';
import type { AnalyzerResult } from '../analyzers/index.js';

export interface JsonReport {
  version: string;
  timestamp: string;
  score: number;
  grade: string;
  metrics: Array<{
    name: string;
    score: number;
    grade: string;
    weight: number;
    summary: string;
    details: Array<{
      file: string;
      message: string;
      line?: number;
      value?: number;
    }>;
  }>;
  warnings: string[];
}

export function renderJsonReport(
  composite: CompositeScore,
  results: AnalyzerResult[],
): string {
  const report: JsonReport = {
    version: '0.1.0',
    timestamp: new Date().toISOString(),
    score: composite.overall,
    grade: composite.subGrade,
    metrics: composite.metrics.map((m) => {
      const result = results.find((r) => r.metric === m.metric);
      return {
        name: m.metric,
        score: m.score,
        grade: m.grade,
        weight: m.weight,
        summary: m.summary,
        details: result?.details ?? [],
      };
    }),
    warnings: composite.warnings,
  };

  return JSON.stringify(report, null, 2);
}
