export interface ParsedFile {
  path: string;
  relativePath: string;
  content: string;
  lines: string[];
  lineCount: number;
  language: string;
  functions: FunctionInfo[];
}

export interface FunctionInfo {
  name: string;
  startLine: number;
  endLine: number;
  lineCount: number;
  body: string;
}
