import type { Analyzer, AnalyzerResult, AnalyzerDetail } from './types.js';
import type { CodeBenchConfig } from '../config/index.js';
import type { ParsedFile } from '../parsers/index.js';

/**
 * Count cyclomatic complexity for a function body.
 * Each decision point adds 1: if, else if, for, while, do, case, catch, &&, ||, ??, ternary.
 * Base complexity is 1.
 */
function computeCyclomaticComplexity(body: string): number {
  let complexity = 1;

  const decisionPatterns = [
    /\bif\s*\(/g,
    /\belse\s+if\s*\(/g,
    /\bfor\s*\(/g,
    /\bwhile\s*\(/g,
    /\bdo\s*\{/g,
    /\bcase\s+/g,
    /\bcatch\s*\(/g,
    /&&/g,
    /\|\|/g,
    /\?\?/g,
    /\?[^?:]*:/g, // ternary (rough)
  ];

  for (const pattern of decisionPatterns) {
    const matches = body.match(pattern);
    if (matches) complexity += matches.length;
  }

  // Subtract double-counted else-if (counted in both if and else if)
  const elseIfMatches = body.match(/\belse\s+if\s*\(/g);
  if (elseIfMatches) complexity -= elseIfMatches.length;

  return complexity;
}

export const cyclomaticAnalyzer: Analyzer = {
  name: 'cyclomatic',

  async analyze(
    files: ParsedFile[],
    config: CodeBenchConfig,
  ): Promise<AnalyzerResult> {
    const threshold = config.thresholds.max_cyclomatic;
    const details: AnalyzerDetail[] = [];
    let totalFunctions = 0;
    let functionsWithinThreshold = 0;

    for (const file of files) {
      for (const fn of file.functions) {
        const complexity = computeCyclomaticComplexity(fn.body);
        totalFunctions++;

        if (complexity <= threshold) {
          functionsWithinThreshold++;
        } else {
          details.push({
            file: file.relativePath,
            message: `${fn.name}() has cyclomatic complexity ${complexity} (threshold: ${threshold})`,
            line: fn.startLine,
            value: complexity,
          });
        }
      }
    }

    if (totalFunctions === 0) {
      return {
        metric: 'cyclomatic',
        score: 100,
        details: [],
        summary: 'No functions found to analyze',
      };
    }

    const score = Math.round((functionsWithinThreshold / totalFunctions) * 100);
    const violations = totalFunctions - functionsWithinThreshold;

    return {
      metric: 'cyclomatic',
      score,
      details: details.sort((a, b) => (b.value ?? 0) - (a.value ?? 0)),
      summary: `${violations} of ${totalFunctions} functions exceed complexity threshold of ${threshold}`,
    };
  },
};
