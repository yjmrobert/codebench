import type { Analyzer, AnalyzerResult, AnalyzerDetail } from './types.js';
import type { CodeBenchConfig } from '../config/index.js';
import type { ParsedFile } from '../parsers/index.js';

export const sizeAnalyzer: Analyzer = {
  name: 'size',

  async analyze(
    files: ParsedFile[],
    config: CodeBenchConfig,
  ): Promise<AnalyzerResult> {
    const maxFileLines = config.thresholds.max_file_lines;
    const maxFunctionLines = config.thresholds.max_function_lines;
    const details: AnalyzerDetail[] = [];
    let filesOverThreshold = 0;
    let functionsOverThreshold = 0;
    let totalFunctions = 0;
    let totalLines = 0;

    for (const file of files) {
      totalLines += file.lineCount;

      if (file.lineCount > maxFileLines) {
        filesOverThreshold++;
        details.push({
          file: file.relativePath,
          message: `File has ${file.lineCount} lines (threshold: ${maxFileLines})`,
          value: file.lineCount,
        });
      }

      for (const fn of file.functions) {
        totalFunctions++;
        if (fn.lineCount > maxFunctionLines) {
          functionsOverThreshold++;
          details.push({
            file: file.relativePath,
            message: `${fn.name}() has ${fn.lineCount} lines (threshold: ${maxFunctionLines})`,
            line: fn.startLine,
            value: fn.lineCount,
          });
        }
      }
    }

    // Score: percentage of files + functions within thresholds
    const totalItems = files.length + totalFunctions;
    const violations = filesOverThreshold + functionsOverThreshold;
    const score =
      totalItems > 0
        ? Math.round(((totalItems - violations) / totalItems) * 100)
        : 100;

    return {
      metric: 'size',
      score,
      details: details.sort((a, b) => (b.value ?? 0) - (a.value ?? 0)),
      summary: `${totalLines} total lines across ${files.length} files. ${filesOverThreshold} oversized files, ${functionsOverThreshold} oversized functions`,
    };
  },
};
