import { readFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import type { Analyzer, AnalyzerResult, AnalyzerDetail } from './types.js';
import type { CodeBenchConfig } from '../config/index.js';
import type { ParsedFile } from '../parsers/index.js';

export const dependencyAnalyzer: Analyzer = {
  name: 'dependencies',

  async analyze(
    _files: ParsedFile[],
    _config: CodeBenchConfig,
    cwd: string,
  ): Promise<AnalyzerResult> {
    const pkgPath = join(cwd, 'package.json');

    if (!existsSync(pkgPath)) {
      return {
        metric: 'dependencies',
        score: 100,
        details: [{ file: 'package.json', message: 'No package.json found — skipping dependency analysis' }],
        summary: 'No package.json found',
      };
    }

    const pkg = JSON.parse(readFileSync(pkgPath, 'utf-8'));
    const deps = pkg.dependencies || {};
    const devDeps = pkg.devDependencies || {};
    const allDeps = { ...deps, ...devDeps };
    const depNames = Object.keys(allDeps);
    const details: AnalyzerDetail[] = [];

    let issues = 0;

    // Check for wildcard or overly loose versions
    for (const [name, version] of Object.entries(allDeps)) {
      const v = version as string;
      if (v === '*' || v === 'latest') {
        issues++;
        details.push({
          file: 'package.json',
          message: `${name} uses unpinned version "${v}"`,
        });
      }
    }

    // Check total dependency count (flag if excessive)
    const totalDeps = Object.keys(deps).length;
    const totalDevDeps = Object.keys(devDeps).length;
    if (totalDeps > 50) {
      issues++;
      details.push({
        file: 'package.json',
        message: `${totalDeps} production dependencies (consider reducing)`,
        value: totalDeps,
      });
    }

    // Check for lock file
    const hasLock =
      existsSync(join(cwd, 'package-lock.json')) ||
      existsSync(join(cwd, 'yarn.lock')) ||
      existsSync(join(cwd, 'pnpm-lock.yaml'));

    if (!hasLock) {
      issues++;
      details.push({
        file: 'package.json',
        message: 'No lock file found — dependency versions may be non-deterministic',
      });
    }

    // Score based on issues ratio
    const maxIssues = Math.max(depNames.length, 1);
    const score = Math.round(Math.max(0, 100 - (issues / maxIssues) * 100));

    return {
      metric: 'dependencies',
      score: Math.min(100, score),
      details,
      summary: `${depNames.length} total dependencies (${totalDeps} prod, ${totalDevDeps} dev), ${issues} issues found`,
    };
  },
};
