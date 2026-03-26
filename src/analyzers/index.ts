import type { Analyzer } from './types.js';
import { coverageAnalyzer } from './coverage.js';
import { cyclomaticAnalyzer } from './cyclomatic.js';
import { cognitiveAnalyzer } from './cognitive.js';
import { duplicationAnalyzer } from './duplication.js';
import { dependencyAnalyzer } from './dependencies.js';
import { sizeAnalyzer } from './size.js';
import { debtAnalyzer } from './debt.js';

export function getAllAnalyzers(): Analyzer[] {
  return [
    coverageAnalyzer,
    cyclomaticAnalyzer,
    cognitiveAnalyzer,
    duplicationAnalyzer,
    dependencyAnalyzer,
    sizeAnalyzer,
    debtAnalyzer,
  ];
}

export type { Analyzer, AnalyzerResult, AnalyzerDetail } from './types.js';
