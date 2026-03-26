package storage

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codebench/codebench/internal/scoring"
	_ "github.com/mattn/go-sqlite3"
)

type GitInfo struct {
	CommitHash string
	Branch     string
}

type RunRecord struct {
	ID           int
	Timestamp    string
	CommitHash   string
	Branch       string
	OverallScore int
	OverallGrade string
}

func GetGitInfo(cwd string) GitInfo {
	info := GitInfo{CommitHash: "unknown", Branch: "unknown"}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err == nil {
		info.CommitHash = strings.TrimSpace(string(out))
	}

	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = cwd
	out, err = cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(out))
		if branch == "" {
			branch = "detached"
		}
		info.Branch = branch
	}

	return info
}

func ensureDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
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
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func SaveRun(dbPath string, score *scoring.CompositeScore, gitInfo GitInfo) (int64, error) {
	db, err := ensureDB(dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO runs (timestamp, commit_hash, branch, overall_score, overall_grade) VALUES (?, ?, ?, ?, ?)`,
		time.Now().UTC().Format(time.RFC3339),
		gitInfo.CommitHash,
		gitInfo.Branch,
		score.Overall,
		string(score.Grade),
	)
	if err != nil {
		return 0, err
	}

	runID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, metric := range score.Metrics {
		_, err = tx.Exec(
			`INSERT INTO metric_scores (run_id, metric, score, grade) VALUES (?, ?, ?, ?)`,
			runID, string(metric.Metric), metric.Score, metric.Grade,
		)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return runID, nil
}

// GetPreviousRun returns the most recent run record and its per-metric scores.
// Returns nil, nil, nil if no previous runs exist.
func GetPreviousRun(dbPath string) (*RunRecord, map[string]int, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, nil, nil
	}

	db, err := ensureDB(dbPath)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	var r RunRecord
	err = db.QueryRow(
		`SELECT id, timestamp, commit_hash, branch, overall_score, overall_grade FROM runs ORDER BY id DESC LIMIT 1`,
	).Scan(&r.ID, &r.Timestamp, &r.CommitHash, &r.Branch, &r.OverallScore, &r.OverallGrade)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	rows, err := db.Query(
		`SELECT metric, score FROM metric_scores WHERE run_id = ?`, r.ID,
	)
	if err != nil {
		return &r, nil, err
	}
	defer rows.Close()

	metricScores := make(map[string]int)
	for rows.Next() {
		var metric string
		var score int
		if err := rows.Scan(&metric, &score); err != nil {
			continue
		}
		metricScores[metric] = score
	}

	return &r, metricScores, nil
}

func GetHistory(dbPath string, limit int) ([]RunRecord, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, nil
	}

	db, err := ensureDB(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT id, timestamp, commit_hash, branch, overall_score, overall_grade FROM runs ORDER BY id DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RunRecord
	for rows.Next() {
		var r RunRecord
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.CommitHash, &r.Branch, &r.OverallScore, &r.OverallGrade); err != nil {
			continue
		}
		records = append(records, r)
	}

	return records, nil
}
