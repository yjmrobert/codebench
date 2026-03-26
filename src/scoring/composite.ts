import type { MetricName, LetterGrade, CodeBenchConfig } from '../config/index.js';
import type { AnalyzerResult } from '../analyzers/index.js';
import { toLetterGrade, toSubGrade } from './grading.js';

export interface MetricScore {
  metric: MetricName;
  score: number;
  grade: string;
  weight: number;
  summary: string;
}

export interface CompositeScore {
  overall: number;
  grade: LetterGrade;
  subGrade: string;
  metrics: MetricScore[];
  warnings: string[];
}

export function computeCompositeScore(
  results: AnalyzerResult[],
  config: CodeBenchConfig,
): CompositeScore {
  const weights = config.weights;
  const totalWeight = Object.values(weights).reduce((s, w) => s + w, 0);

  let weightedSum = 0;
  const metrics: MetricScore[] = [];
  const warnings: string[] = [];

  for (const result of results) {
    const weight = weights[result.metric] ?? 0;
    const normalizedWeight = totalWeight > 0 ? weight / totalWeight : 0;
    weightedSum += result.score * normalizedWeight;

    metrics.push({
      metric: result.metric,
      score: result.score,
      grade: toSubGrade(result.score),
      weight,
      summary: result.summary,
    });

    // Add top warnings from each metric
    for (const detail of result.details.slice(0, 3)) {
      if (result.score < 70) {
        warnings.push(detail.message);
      }
    }
  }

  const overall = Math.round(weightedSum);

  return {
    overall,
    grade: toLetterGrade(overall),
    subGrade: toSubGrade(overall),
    metrics,
    warnings: warnings.slice(0, 10),
  };
}
