import Database from 'better-sqlite3';
import { mkdirSync, existsSync } from 'node:fs';
import { dirname } from 'node:path';
import type { CompositeScore } from '../scoring/index.js';
import type { GitInfo } from './git-info.js';

export interface RunRecord {
  id: number;
  timestamp: string;
  commitHash: string;
  branch: string;
  overallScore: number;
  overallGrade: string;
}

export interface MetricRecord {
  metric: string;
  score: number;
  grade: string;
}

function ensureDb(dbPath: string): Database.Database {
  const dir = dirname(dbPath);
  if (!existsSync(dir)) {
    mkdirSync(dir, { recursive: true });
  }

  const db = new Database(dbPath);
  db.pragma('journal_mode = WAL');

  db.exec(`
    CREATE TABLE IF NOT EXISTS runs (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      timestamp TEXT NOT NULL,
      commit_hash TEXT NOT NULL,
      branch TEXT NOT NULL,
      overall_score REAL NOT NULL,
      overall_grade TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS metric_scores (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      run_id INTEGER NOT NULL REFERENCES runs(id),
      metric TEXT NOT NULL,
      score REAL NOT NULL,
      grade TEXT NOT NULL
    );
  `);

  return db;
}

export function saveRun(
  dbPath: string,
  score: CompositeScore,
  gitInfo: GitInfo,
): number {
  const db = ensureDb(dbPath);

  try {
    const insertRun = db.prepare(`
      INSERT INTO runs (timestamp, commit_hash, branch, overall_score, overall_grade)
      VALUES (?, ?, ?, ?, ?)
    `);

    const insertMetric = db.prepare(`
      INSERT INTO metric_scores (run_id, metric, score, grade)
      VALUES (?, ?, ?, ?)
    `);

    const result = insertRun.run(
      new Date().toISOString(),
      gitInfo.commitHash,
      gitInfo.branch,
      score.overall,
      score.grade,
    );

    const runId = result.lastInsertRowid as number;

    for (const metric of score.metrics) {
      insertMetric.run(runId, metric.metric, metric.score, metric.grade);
    }

    return runId;
  } finally {
    db.close();
  }
}

export function getHistory(dbPath: string, limit = 20): RunRecord[] {
  if (!existsSync(dbPath)) return [];

  const db = ensureDb(dbPath);

  try {
    const rows = db
      .prepare(
        `SELECT id, timestamp, commit_hash, branch, overall_score, overall_grade
       FROM runs ORDER BY id DESC LIMIT ?`,
      )
      .all(limit) as Array<{
      id: number;
      timestamp: string;
      commit_hash: string;
      branch: string;
      overall_score: number;
      overall_grade: string;
    }>;

    return rows.map((r) => ({
      id: r.id,
      timestamp: r.timestamp,
      commitHash: r.commit_hash,
      branch: r.branch,
      overallScore: r.overall_score,
      overallGrade: r.overall_grade,
    }));
  } finally {
    db.close();
  }
}

export function getMetricHistory(
  dbPath: string,
  runId: number,
): MetricRecord[] {
  if (!existsSync(dbPath)) return [];

  const db = ensureDb(dbPath);

  try {
    return db
      .prepare(
        `SELECT metric, score, grade FROM metric_scores WHERE run_id = ?`,
      )
      .all(runId) as MetricRecord[];
  } finally {
    db.close();
  }
}
