import type { Analyzer, AnalyzerResult, AnalyzerDetail } from './types.js';
import type { CodeBenchConfig } from '../config/index.js';
import type { ParsedFile } from '../parsers/index.js';

const MIN_DUPLICATE_LINES = 6;

/**
 * Normalize a line for comparison: trim whitespace, collapse spaces.
 * Skip empty lines and single-character lines (braces, etc.)
 */
function normalizeLine(line: string): string | null {
  const trimmed = line.trim();
  if (trimmed.length <= 1) return null;
  // Collapse whitespace
  return trimmed.replace(/\s+/g, ' ');
}

/**
 * Line-based duplication detection.
 * Hash sliding windows of N consecutive normalized lines, find collisions across files.
 */
export const duplicationAnalyzer: Analyzer = {
  name: 'duplication',

  async analyze(
    files: ParsedFile[],
    config: CodeBenchConfig,
  ): Promise<AnalyzerResult> {
    const windowSize = MIN_DUPLICATE_LINES;
    const details: AnalyzerDetail[] = [];

    // Build normalized line sequences per file
    interface FileLines {
      file: ParsedFile;
      normalizedLines: { text: string; originalLine: number }[];
    }

    const fileLines: FileLines[] = [];
    for (const file of files) {
      const normalized: { text: string; originalLine: number }[] = [];
      for (let i = 0; i < file.lines.length; i++) {
        const n = normalizeLine(file.lines[i]);
        if (n) normalized.push({ text: n, originalLine: i + 1 });
      }
      fileLines.push({ file, normalizedLines: normalized });
    }

    // Hash all windows and detect duplicates
    const windowMap = new Map<string, { file: string; line: number }[]>();
    let totalNormalizedLines = 0;

    for (const fl of fileLines) {
      totalNormalizedLines += fl.normalizedLines.length;
      for (let i = 0; i <= fl.normalizedLines.length - windowSize; i++) {
        const window = fl.normalizedLines
          .slice(i, i + windowSize)
          .map((l) => l.text)
          .join('\n');

        const existing = windowMap.get(window);
        if (existing) {
          existing.push({
            file: fl.file.relativePath,
            line: fl.normalizedLines[i].originalLine,
          });
        } else {
          windowMap.set(window, [
            {
              file: fl.file.relativePath,
              line: fl.normalizedLines[i].originalLine,
            },
          ]);
        }
      }
    }

    // Count duplicated lines
    let duplicatedLines = 0;
    const seenDuplicates = new Set<string>();

    for (const [, locations] of windowMap) {
      if (locations.length <= 1) continue;

      // Count each duplicate window once per file location
      for (const loc of locations) {
        const key = `${loc.file}:${loc.line}`;
        if (!seenDuplicates.has(key)) {
          seenDuplicates.add(key);
          duplicatedLines += windowSize;
        }
      }

      // Report first few duplicates
      if (details.length < 20) {
        const first = locations[0];
        const others = locations.slice(1);
        for (const other of others) {
          if (details.length >= 20) break;
          details.push({
            file: first.file,
            message: `Lines ${first.line}-${first.line + windowSize - 1} duplicated in ${other.file}:${other.line}`,
            line: first.line,
          });
        }
      }
    }

    const duplicationPct =
      totalNormalizedLines > 0
        ? (duplicatedLines / totalNormalizedLines) * 100
        : 0;

    // Score is inverse of duplication percentage
    // 0% duplication = 100 score, max_duplication_pct% or more = 0
    const maxPct = config.thresholds.max_duplication_pct;
    const score = Math.round(Math.max(0, Math.min(100, 100 - (duplicationPct / maxPct) * 100)));

    return {
      metric: 'duplication',
      score,
      details,
      summary: `${duplicationPct.toFixed(1)}% code duplication detected (${duplicatedLines} duplicated lines)`,
    };
  },
};
