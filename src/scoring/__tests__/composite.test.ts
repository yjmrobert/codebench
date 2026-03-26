import { describe, it, expect } from 'vitest';
import { computeCompositeScore } from '../composite.js';
import { DEFAULT_CONFIG } from '../../config/index.js';
import type { AnalyzerResult } from '../../analyzers/types.js';

function makeResult(metric: string, score: number): AnalyzerResult {
  return {
    metric: metric as any,
    score,
    details: [],
    summary: `${metric}: ${score}`,
  };
}

describe('computeCompositeScore', () => {
  it('computes weighted average', () => {
    const results: AnalyzerResult[] = [
      makeResult('coverage', 80),
      makeResult('cyclomatic', 90),
      makeResult('cognitive', 85),
      makeResult('duplication', 95),
      makeResult('dependencies', 70),
      makeResult('size', 88),
      makeResult('debt', 75),
    ];

    const composite = computeCompositeScore(results, DEFAULT_CONFIG);

    // Weighted: (80*25 + 90*15 + 85*15 + 95*15 + 70*10 + 88*10 + 75*10) / 100
    // = (2000 + 1350 + 1275 + 1425 + 700 + 880 + 750) / 100
    // = 8380 / 100 = 83.8 -> 84
    expect(composite.overall).toBe(84);
    expect(composite.grade).toBe('B');
  });

  it('returns 100 when all scores are 100', () => {
    const results = [
      'coverage', 'cyclomatic', 'cognitive', 'duplication',
      'dependencies', 'size', 'debt',
    ].map((m) => makeResult(m, 100));

    const composite = computeCompositeScore(results, DEFAULT_CONFIG);
    expect(composite.overall).toBe(100);
    expect(composite.grade).toBe('A');
  });

  it('returns 0 when all scores are 0', () => {
    const results = [
      'coverage', 'cyclomatic', 'cognitive', 'duplication',
      'dependencies', 'size', 'debt',
    ].map((m) => makeResult(m, 0));

    const composite = computeCompositeScore(results, DEFAULT_CONFIG);
    expect(composite.overall).toBe(0);
    expect(composite.grade).toBe('F');
  });
});
