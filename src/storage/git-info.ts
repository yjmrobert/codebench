import { execSync } from 'node:child_process';

export interface GitInfo {
  commitHash: string;
  branch: string;
}

export function getGitInfo(cwd: string): GitInfo {
  try {
    const commitHash = execSync('git rev-parse HEAD', { cwd, encoding: 'utf-8' }).trim();
    const branch = execSync('git branch --show-current', { cwd, encoding: 'utf-8' }).trim();
    return { commitHash, branch: branch || 'detached' };
  } catch {
    return { commitHash: 'unknown', branch: 'unknown' };
  }
}
