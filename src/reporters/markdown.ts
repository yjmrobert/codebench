import type { CompositeScore } from '../scoring/index.js';
import type { AnalyzerResult } from '../analyzers/index.js';

const METRIC_LABELS: Record<string, string> = {
  coverage: 'Coverage',
  cyclomatic: 'Cyclomatic Complexity',
  cognitive: 'Cognitive Complexity',
  duplication: 'Code Duplication',
  dependencies: 'Dependency Health',
  size: 'Code Size & Structure',
  debt: 'Technical Debt',
};

export function renderMarkdownReport(
  composite: CompositeScore,
  results: AnalyzerResult[],
  projectName: string,
): string {
  const lines: string[] = [];

  lines.push(`## CodeBench Report — ${projectName}`);
  lines.push('');
  lines.push(`**Score: ${composite.overall} / 100 (${composite.subGrade})**`);
  lines.push('');
  lines.push('| Metric | Score | Grade |');
  lines.push('|--------|-------|-------|');

  for (const metric of composite.metrics) {
    const label = METRIC_LABELS[metric.metric] ?? metric.metric;
    lines.push(`| ${label} | ${metric.score}% | ${metric.grade} |`);
  }

  if (composite.warnings.length > 0) {
    lines.push('');
    lines.push('### Warnings');
    lines.push('');
    for (const warning of composite.warnings) {
      lines.push(`- ⚠️ ${warning}`);
    }
  }

  lines.push('');
  return lines.join('\n');
}
