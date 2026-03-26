import { getHistory } from '../../storage/index.js';
import { renderHistoryReport } from '../../reporters/index.js';

export interface HistoryOptions {
  limit?: number;
}

export async function runHistory(
  cwd: string,
  options: HistoryOptions,
): Promise<number> {
  const dbPath = `${cwd}/.codebench/history.db`;
  const runs = getHistory(dbPath, options.limit ?? 20);
  console.log(renderHistoryReport(runs));
  return 0;
}
