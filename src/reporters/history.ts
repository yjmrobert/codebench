import chalk from 'chalk';
import Table from 'cli-table3';
import type { RunRecord } from '../storage/index.js';

function gradeColor(grade: string): (text: string) => string {
  if (grade.startsWith('A')) return chalk.green;
  if (grade.startsWith('B')) return chalk.cyan;
  if (grade.startsWith('C')) return chalk.yellow;
  if (grade.startsWith('D')) return chalk.red;
  return chalk.bgRed.white;
}

function trendArrow(current: number, previous: number | null): string {
  if (previous === null) return ' ';
  if (current > previous) return chalk.green('▲');
  if (current < previous) return chalk.red('▼');
  return chalk.gray('─');
}

export function renderHistoryReport(runs: RunRecord[]): string {
  if (runs.length === 0) {
    return chalk.yellow('No history found. Run `codebench` to record your first score.\n');
  }

  const table = new Table({
    head: ['Date', 'Branch', 'Commit', 'Score', 'Grade', 'Trend'],
    style: { head: ['cyan'] },
  });

  // Runs are in reverse chronological order
  const reversed = [...runs].reverse();
  for (let i = 0; i < reversed.length; i++) {
    const run = reversed[i];
    const prev = i > 0 ? reversed[i - 1].overallScore : null;
    const date = new Date(run.timestamp).toLocaleDateString();
    const color = gradeColor(run.overallGrade);

    table.push([
      date,
      run.branch,
      run.commitHash.substring(0, 7),
      color(run.overallScore.toString()),
      color(run.overallGrade),
      trendArrow(run.overallScore, prev),
    ]);
  }

  return `\n${chalk.bold('CodeBench History')}\n${table.toString()}\n`;
}
