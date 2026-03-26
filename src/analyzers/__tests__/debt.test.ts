import { describe, it, expect } from 'vitest';
import { debtAnalyzer } from '../debt.js';
import { DEFAULT_CONFIG } from '../../config/index.js';
import type { ParsedFile } from '../../parsers/types.js';

function makeFile(content: string): ParsedFile {
  const lines = content.split('\n');
  return {
    path: '/test/file.ts',
    relativePath: 'file.ts',
    content,
    lines,
    lineCount: lines.length,
    language: 'typescript',
    functions: [],
  };
}

describe('debtAnalyzer', () => {
  it('finds TODO and FIXME markers', async () => {
    const file = makeFile(`
      // TODO: fix this later
      const x = 1;
      // FIXME: broken
      // HACK: temporary workaround
    `);

    const result = await debtAnalyzer.analyze([file], DEFAULT_CONFIG, '/test');
    expect(result.details.length).toBe(3);
    expect(result.score).toBeLessThan(100);
  });

  it('returns 100 for clean code', async () => {
    const file = makeFile(`
      const x = 1;
      const y = 2;
      export function add(a: number, b: number) { return a + b; }
    `);

    const result = await debtAnalyzer.analyze([file], DEFAULT_CONFIG, '/test');
    expect(result.details.length).toBe(0);
    expect(result.score).toBe(100);
  });
});
