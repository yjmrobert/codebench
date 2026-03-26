import type { Analyzer, AnalyzerResult, AnalyzerDetail } from './types.js';
import type { CodeBenchConfig } from '../config/index.js';
import type { ParsedFile } from '../parsers/index.js';

/**
 * Compute cognitive complexity using a simplified Sonar model.
 * Increments for control flow breaks, with nesting depth as multiplier.
 */
function computeCognitiveComplexity(body: string): number {
  let complexity = 0;
  let nestingDepth = 0;
  const lines = body.split('\n');

  for (const line of lines) {
    const trimmed = line.trim();

    // Track nesting via braces (simplified)
    const opens = (trimmed.match(/\{/g) || []).length;
    const closes = (trimmed.match(/\}/g) || []).length;

    // Increment for control flow keywords (with nesting penalty)
    if (/\b(if|else if)\s*\(/.test(trimmed)) {
      complexity += 1 + nestingDepth;
    } else if (/\belse\s*\{/.test(trimmed)) {
      complexity += 1;
    }

    if (/\b(for|while)\s*\(/.test(trimmed)) {
      complexity += 1 + nestingDepth;
    }

    if (/\bdo\s*\{/.test(trimmed)) {
      complexity += 1 + nestingDepth;
    }

    if (/\bswitch\s*\(/.test(trimmed)) {
      complexity += 1 + nestingDepth;
    }

    if (/\bcatch\s*\(/.test(trimmed)) {
      complexity += 1 + nestingDepth;
    }

    // Logical operators add flat increment
    const logicalOps = (trimmed.match(/&&|\|\||\?\?/g) || []).length;
    complexity += logicalOps;

    // Ternary
    if (/\?[^?:]*:/.test(trimmed)) {
      complexity += 1 + nestingDepth;
    }

    // Recursion detection (calling own function name) - skip for simplicity in v0.1

    nestingDepth += opens - closes;
    if (nestingDepth < 0) nestingDepth = 0;
  }

  return complexity;
}

export const cognitiveAnalyzer: Analyzer = {
  name: 'cognitive',

  async analyze(
    files: ParsedFile[],
    config: CodeBenchConfig,
  ): Promise<AnalyzerResult> {
    const threshold = config.thresholds.max_cognitive;
    const details: AnalyzerDetail[] = [];
    let totalFunctions = 0;
    let functionsWithinThreshold = 0;

    for (const file of files) {
      for (const fn of file.functions) {
        const complexity = computeCognitiveComplexity(fn.body);
        totalFunctions++;

        if (complexity <= threshold) {
          functionsWithinThreshold++;
        } else {
          details.push({
            file: file.relativePath,
            message: `${fn.name}() has cognitive complexity ${complexity} (threshold: ${threshold})`,
            line: fn.startLine,
            value: complexity,
          });
        }
      }
    }

    if (totalFunctions === 0) {
      return {
        metric: 'cognitive',
        score: 100,
        details: [],
        summary: 'No functions found to analyze',
      };
    }

    const score = Math.round((functionsWithinThreshold / totalFunctions) * 100);
    const violations = totalFunctions - functionsWithinThreshold;

    return {
      metric: 'cognitive',
      score,
      details: details.sort((a, b) => (b.value ?? 0) - (a.value ?? 0)),
      summary: `${violations} of ${totalFunctions} functions exceed cognitive complexity threshold of ${threshold}`,
    };
  },
};
