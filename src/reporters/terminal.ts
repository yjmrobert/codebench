import chalk from 'chalk';
import Table from 'cli-table3';
import type { CompositeScore, MetricScore } from '../scoring/index.js';
import type { AnalyzerResult } from '../analyzers/index.js';

const METRIC_LABELS: Record<string, string> = {
  coverage: 'Coverage',
  cyclomatic: 'Cyclomatic',
  cognitive: 'Cognitive',
  duplication: 'Duplication',
  dependencies: 'Dependencies',
  size: 'Size/Structure',
  debt: 'Tech Debt',
};

function gradeColor(grade: string): (text: string) => string {
  if (grade.startsWith('A')) return chalk.green;
  if (grade.startsWith('B')) return chalk.cyan;
  if (grade.startsWith('C')) return chalk.yellow;
  if (grade.startsWith('D')) return chalk.red;
  return chalk.bgRed.white;
}

function progressBar(score: number, width = 10): string {
  const filled = Math.round((score / 100) * width);
  const empty = width - filled;
  return chalk.green('█'.repeat(filled)) + chalk.gray('░'.repeat(empty));
}

export function renderTerminalReport(
  composite: CompositeScore,
  results: AnalyzerResult[],
  projectName: string,
): string {
  const lines: string[] = [];
  const color = gradeColor(composite.grade);

  // Header
  lines.push('');
  lines.push(
    chalk.bold(
      `┌${'─'.repeat(47)}┐`,
    ),
  );
  lines.push(
    chalk.bold(
      `│  CodeBench Report — ${projectName.padEnd(25)}│`,
    ),
  );
  lines.push(
    chalk.bold(
      `│  Score: ${color(`${composite.overall} / 100`)}  (${color(composite.subGrade)})${' '.repeat(Math.max(0, 25 - composite.overall.toString().length - composite.subGrade.length))}│`,
    ),
  );
  lines.push(
    chalk.bold(
      `├${'─'.repeat(47)}┤`,
    ),
  );

  // Metric rows
  for (const metric of composite.metrics) {
    const label = (METRIC_LABELS[metric.metric] ?? metric.metric).padEnd(18);
    const bar = progressBar(metric.score);
    const pct = `${metric.score}%`.padStart(4);
    const grade = gradeColor(metric.grade)(`(${metric.grade})`);
    lines.push(
      chalk.bold(`│  `) +
        `${label} ${bar}  ${pct}  ${grade}` +
        ' '.repeat(Math.max(0, 3)) +
        chalk.bold('│'),
    );
  }

  // Warnings
  if (composite.warnings.length > 0) {
    lines.push(
      chalk.bold(
        `├${'─'.repeat(47)}┤`,
      ),
    );
    for (const warning of composite.warnings.slice(0, 5)) {
      const truncated = warning.length > 43 ? warning.substring(0, 40) + '...' : warning;
      lines.push(
        chalk.bold('│  ') + chalk.yellow('⚠  ') + truncated,
      );
    }
  }

  lines.push(
    chalk.bold(
      `└${'─'.repeat(47)}┘`,
    ),
  );
  lines.push('');

  return lines.join('\n');
}
