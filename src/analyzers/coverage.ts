import { readFileSync, existsSync } from 'node:fs';
import { resolve } from 'node:path';
import type { Analyzer, AnalyzerResult } from './types.js';
import type { CodeBenchConfig } from '../config/index.js';
import type { ParsedFile } from '../parsers/index.js';

interface FileCoverage {
  file: string;
  linesFound: number;
  linesHit: number;
}

function parseLcov(content: string): FileCoverage[] {
  const files: FileCoverage[] = [];
  let current: Partial<FileCoverage> = {};

  for (const line of content.split('\n')) {
    const trimmed = line.trim();
    if (trimmed.startsWith('SF:')) {
      current = { file: trimmed.slice(3) };
    } else if (trimmed.startsWith('LF:')) {
      current.linesFound = parseInt(trimmed.slice(3), 10);
    } else if (trimmed.startsWith('LH:')) {
      current.linesHit = parseInt(trimmed.slice(3), 10);
    } else if (trimmed === 'end_of_record') {
      if (current.file && current.linesFound !== undefined && current.linesHit !== undefined) {
        files.push(current as FileCoverage);
      }
      current = {};
    }
  }

  return files;
}

export const coverageAnalyzer: Analyzer = {
  name: 'coverage',

  async analyze(
    _files: ParsedFile[],
    config: CodeBenchConfig,
    cwd: string,
  ): Promise<AnalyzerResult> {
    const reportPath = resolve(cwd, config.coverage.report_path);

    if (!existsSync(reportPath)) {
      return {
        metric: 'coverage',
        score: 0,
        details: [
          {
            file: reportPath,
            message: `Coverage report not found at ${config.coverage.report_path}. Run your test suite with coverage enabled.`,
          },
        ],
        summary: 'No coverage report found',
      };
    }

    const content = readFileSync(reportPath, 'utf-8');
    const fileCoverages = parseLcov(content);

    if (fileCoverages.length === 0) {
      return {
        metric: 'coverage',
        score: 0,
        details: [{ file: reportPath, message: 'Coverage report is empty' }],
        summary: '0% coverage (empty report)',
      };
    }

    const totalLines = fileCoverages.reduce((sum, f) => sum + f.linesFound, 0);
    const totalHit = fileCoverages.reduce((sum, f) => sum + f.linesHit, 0);
    const percentage = totalLines > 0 ? (totalHit / totalLines) * 100 : 0;

    const details = fileCoverages
      .filter((f) => f.linesFound > 0)
      .map((f) => {
        const pct = ((f.linesHit / f.linesFound) * 100).toFixed(1);
        return {
          file: f.file,
          message: `${pct}% coverage (${f.linesHit}/${f.linesFound} lines)`,
          value: parseFloat(pct),
        };
      })
      .sort((a, b) => (a.value ?? 0) - (b.value ?? 0));

    return {
      metric: 'coverage',
      score: Math.round(percentage),
      details,
      summary: `${percentage.toFixed(1)}% line coverage across ${fileCoverages.length} files`,
    };
  },
};
