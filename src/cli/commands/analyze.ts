import { basename } from 'node:path';
import type { MetricName, CodeBenchConfig } from '../../config/index.js';
import { loadConfig } from '../../config/index.js';
import { discoverFiles, parseFiles } from '../../parsers/index.js';
import { getAllAnalyzers } from '../../analyzers/index.js';
import type { AnalyzerResult } from '../../analyzers/index.js';
import { computeCompositeScore } from '../../scoring/index.js';
import { saveRun, getGitInfo } from '../../storage/index.js';
import {
  renderTerminalReport,
  renderJsonReport,
  renderMarkdownReport,
} from '../../reporters/index.js';

export interface AnalyzeOptions {
  json?: boolean;
  format?: 'terminal' | 'json' | 'markdown';
  metric?: string;
  threshold?: number;
  config?: string;
  noCache?: boolean;
}

export async function runAnalysis(
  cwd: string,
  options: AnalyzeOptions,
): Promise<number> {
  const config = loadConfig(cwd, options.config);
  const format = options.json ? 'json' : options.format ?? 'terminal';

  // Discover and parse files
  const filePaths = discoverFiles(config, cwd);

  if (filePaths.length === 0 && format === 'terminal') {
    console.error(
      'No source files found. Check your language settings and ignore patterns.',
    );
    return 2;
  }

  const files = parseFiles(filePaths, cwd);

  // Run analyzers
  const analyzers = getAllAnalyzers();
  const targetAnalyzers = options.metric
    ? analyzers.filter((a) => a.name === options.metric)
    : analyzers;

  if (options.metric && targetAnalyzers.length === 0) {
    console.error(`Unknown metric: ${options.metric}`);
    return 2;
  }

  const results: AnalyzerResult[] = [];
  for (const analyzer of targetAnalyzers) {
    const result = await analyzer.analyze(files, config, cwd);
    results.push(result);
  }

  // Compute composite score
  const composite = computeCompositeScore(results, config);

  // Save to history
  const dbPath = `${cwd}/.codebench/history.db`;
  try {
    const gitInfo = getGitInfo(cwd);
    saveRun(dbPath, composite, gitInfo);
  } catch {
    // History save failure is non-fatal
  }

  // Render output
  const projectName = basename(cwd);
  switch (format) {
    case 'json':
      console.log(renderJsonReport(composite, results));
      break;
    case 'markdown':
      console.log(renderMarkdownReport(composite, results, projectName));
      break;
    default:
      console.log(renderTerminalReport(composite, results, projectName));
      break;
  }

  // Check threshold
  if (options.threshold !== undefined && composite.overall < options.threshold) {
    if (format === 'terminal') {
      console.error(
        `Score ${composite.overall} is below threshold ${options.threshold}`,
      );
    }
    return 1;
  }

  return 0;
}
