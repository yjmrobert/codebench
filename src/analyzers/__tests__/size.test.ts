import { describe, it, expect } from 'vitest';
import { sizeAnalyzer } from '../size.js';
import { DEFAULT_CONFIG } from '../../config/index.js';
import type { ParsedFile } from '../../parsers/types.js';

function makeFile(lineCount: number, functions: { name: string; lineCount: number }[] = []): ParsedFile {
  const lines = Array(lineCount).fill('const x = 1;');
  return {
    path: '/test/file.ts',
    relativePath: 'file.ts',
    content: lines.join('\n'),
    lines,
    lineCount,
    language: 'typescript',
    functions: functions.map((f, i) => ({
      name: f.name,
      startLine: i * f.lineCount + 1,
      endLine: (i + 1) * f.lineCount,
      lineCount: f.lineCount,
      body: Array(f.lineCount).fill('  const x = 1;').join('\n'),
    })),
  };
}

describe('sizeAnalyzer', () => {
  it('scores 100 for small files with small functions', async () => {
    const file = makeFile(50, [{ name: 'foo', lineCount: 10 }]);
    const result = await sizeAnalyzer.analyze([file], DEFAULT_CONFIG, '/test');
    expect(result.score).toBe(100);
  });

  it('penalizes oversized files', async () => {
    const file = makeFile(500, []);
    const result = await sizeAnalyzer.analyze([file], DEFAULT_CONFIG, '/test');
    expect(result.score).toBe(0);
    expect(result.details.length).toBe(1);
  });

  it('penalizes oversized functions', async () => {
    const file = makeFile(100, [{ name: 'bigFunc', lineCount: 80 }]);
    const result = await sizeAnalyzer.analyze([file], DEFAULT_CONFIG, '/test');
    expect(result.score).toBe(50); // 1 of 2 items (file + function) over threshold
  });
});
