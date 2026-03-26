import { readFileSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import yaml from 'js-yaml';
import { DEFAULT_CONFIG } from './defaults.js';
import type { CodeBenchConfig } from './schema.js';

export function loadConfig(cwd: string, configPath?: string): CodeBenchConfig {
  const filePath = configPath ?? join(cwd, '.codebench.yml');

  if (!existsSync(filePath)) {
    return { ...DEFAULT_CONFIG };
  }

  const raw = readFileSync(filePath, 'utf-8');
  const parsed = yaml.load(raw) as Partial<CodeBenchConfig> | null;

  if (!parsed || typeof parsed !== 'object') {
    return { ...DEFAULT_CONFIG };
  }

  return {
    languages: parsed.languages ?? DEFAULT_CONFIG.languages,
    weights: { ...DEFAULT_CONFIG.weights, ...parsed.weights },
    thresholds: { ...DEFAULT_CONFIG.thresholds, ...parsed.thresholds },
    ignore: parsed.ignore ?? DEFAULT_CONFIG.ignore,
    coverage: { ...DEFAULT_CONFIG.coverage, ...parsed.coverage },
    ci: { ...DEFAULT_CONFIG.ci, ...parsed.ci },
  };
}

export { DEFAULT_CONFIG } from './defaults.js';
