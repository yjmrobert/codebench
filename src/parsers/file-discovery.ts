import { globSync } from 'glob';
import { join } from 'node:path';
import type { CodeBenchConfig } from '../config/index.js';

const LANGUAGE_EXTENSIONS: Record<string, string[]> = {
  javascript: ['js', 'jsx', 'mjs', 'cjs'],
  typescript: ['ts', 'tsx', 'mts', 'cts'],
  python: ['py'],
  go: ['go'],
  rust: ['rs'],
};

export function discoverFiles(config: CodeBenchConfig, cwd: string): string[] {
  const extensions: string[] = [];
  for (const lang of config.languages) {
    const exts = LANGUAGE_EXTENSIONS[lang];
    if (exts) extensions.push(...exts);
  }

  if (extensions.length === 0) return [];

  const pattern =
    extensions.length === 1
      ? `**/*.${extensions[0]}`
      : `**/*.{${extensions.join(',')}}`;

  const ignorePatterns = config.ignore.map((p) =>
    p.startsWith('/') ? p : join(cwd, p),
  );

  return globSync(pattern, {
    cwd,
    absolute: true,
    ignore: config.ignore,
    nodir: true,
  });
}
