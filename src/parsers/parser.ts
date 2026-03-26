import { readFileSync } from 'node:fs';
import { relative, extname } from 'node:path';
import type { ParsedFile, FunctionInfo } from './types.js';

const EXT_TO_LANGUAGE: Record<string, string> = {
  '.js': 'javascript',
  '.jsx': 'javascript',
  '.mjs': 'javascript',
  '.cjs': 'javascript',
  '.ts': 'typescript',
  '.tsx': 'typescript',
  '.mts': 'typescript',
  '.cts': 'typescript',
  '.py': 'python',
  '.go': 'go',
  '.rs': 'rust',
};

/**
 * Extract functions from JS/TS source using brace-matching heuristic.
 * Handles: function declarations, arrow functions assigned to const/let/var,
 * class methods, and exported functions.
 */
function extractFunctions(content: string, lines: string[]): FunctionInfo[] {
  const functions: FunctionInfo[] = [];

  // Pattern matches function declarations and expressions
  const patterns = [
    // function name(...) {
    /(?:export\s+(?:default\s+)?)?(?:async\s+)?function\s+(\w+)\s*\(/g,
    // const name = (...) => {  or  const name = function
    /(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?(?:\([^)]*\)|[^=])\s*=>\s*\{/g,
    /(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?function\s*\(/g,
    // class methods: name(...) {  (but not if, for, while, etc.)
    /^\s+(?:async\s+)?(?:static\s+)?(\w+)\s*\([^)]*\)\s*\{/gm,
  ];

  for (const pattern of patterns) {
    let match;
    while ((match = pattern.exec(content)) !== null) {
      const name = match[1];
      if (['if', 'for', 'while', 'switch', 'catch', 'else'].includes(name)) continue;

      const startOffset = match.index;
      const startLine = content.substring(0, startOffset).split('\n').length;

      // Find the opening brace after the match
      let braceStart = content.indexOf('{', startOffset + match[0].length - 1);
      if (braceStart === -1) braceStart = content.indexOf('{', startOffset);
      if (braceStart === -1) continue;

      // Brace-match to find the end
      let depth = 0;
      let endOffset = braceStart;
      for (let i = braceStart; i < content.length; i++) {
        if (content[i] === '{') depth++;
        else if (content[i] === '}') {
          depth--;
          if (depth === 0) {
            endOffset = i;
            break;
          }
        }
      }

      const endLine = content.substring(0, endOffset).split('\n').length;
      const lineCount = endLine - startLine + 1;
      const body = lines.slice(startLine - 1, endLine).join('\n');

      // Avoid duplicates at the same start line
      if (!functions.some((f) => f.startLine === startLine)) {
        functions.push({ name, startLine, endLine, lineCount, body });
      }
    }
  }

  return functions.sort((a, b) => a.startLine - b.startLine);
}

export function parseFile(filePath: string, cwd: string): ParsedFile {
  const content = readFileSync(filePath, 'utf-8');
  const lines = content.split('\n');
  const ext = extname(filePath);
  const language = EXT_TO_LANGUAGE[ext] ?? 'unknown';

  const functions =
    language === 'javascript' || language === 'typescript'
      ? extractFunctions(content, lines)
      : [];

  return {
    path: filePath,
    relativePath: relative(cwd, filePath),
    content,
    lines,
    lineCount: lines.length,
    language,
    functions,
  };
}

export function parseFiles(filePaths: string[], cwd: string): ParsedFile[] {
  return filePaths.map((fp) => parseFile(fp, cwd));
}
