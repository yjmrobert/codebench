import type { Analyzer, AnalyzerResult, AnalyzerDetail } from './types.js';
import type { CodeBenchConfig } from '../config/index.js';
import type { ParsedFile } from '../parsers/index.js';

const DEBT_MARKERS = ['TODO', 'FIXME', 'HACK', 'XXX'];
const MARKER_PATTERN = new RegExp(`\\b(${DEBT_MARKERS.join('|')})\\b`, 'gi');

export const debtAnalyzer: Analyzer = {
  name: 'debt',

  async analyze(
    files: ParsedFile[],
    config: CodeBenchConfig,
  ): Promise<AnalyzerResult> {
    const details: AnalyzerDetail[] = [];
    let totalMarkers = 0;
    let totalLines = 0;

    for (const file of files) {
      totalLines += file.lineCount;

      for (let i = 0; i < file.lines.length; i++) {
        const line = file.lines[i];
        const matches = line.match(MARKER_PATTERN);
        if (matches) {
          totalMarkers += matches.length;
          for (const match of matches) {
            details.push({
              file: file.relativePath,
              message: `${match.toUpperCase()}: ${line.trim().substring(0, 100)}`,
              line: i + 1,
            });
          }
        }
      }
    }

    // Density = markers per 1000 LOC
    const density = totalLines > 0 ? (totalMarkers / totalLines) * 1000 : 0;
    const maxDensity = config.thresholds.max_debt_density;

    // Score: inverse of density ratio
    const score = Math.round(Math.max(0, Math.min(100, 100 - (density / maxDensity) * 100)));

    return {
      metric: 'debt',
      score,
      details: details.slice(0, 50), // cap details at 50
      summary: `${totalMarkers} debt markers found (density: ${density.toFixed(1)} per 1K LOC)`,
    };
  },
};
