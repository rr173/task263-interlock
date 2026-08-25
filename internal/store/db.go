// Package store 提供 SQLite 持久化。
//
// 使用 modernc.org/sqlite（纯 Go，CGO 无关），单文件数据库。
// 所有实体通过 *_store.go 中的 Store 类型访问，迁移在 Open 时自动执行。
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB 持有 SQLite 连接。
type DB struct {
	*sql.DB
	path string
}

// Open 打开（或创建）SQLite 数据库并执行迁移。
func Open(path string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db := &DB{DB: sqldb, path: path}
	if err := db.migrate(); err != nil {
		_ = sqldb.Close()
		return nil, err
	}
	return db, nil
}

// Path 返回数据库文件路径。
func (d *DB) Path() string { return d.path }

// migrate 执行幂等建表迁移。
func (d *DB) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS segments (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			length_m REAL NOT NULL,
			state TEXT NOT NULL DEFAULT 'clear',
			line_name TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS switches (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			position TEXT NOT NULL,
			normal_to TEXT NOT NULL,
			reverse_to TEXT NOT NULL,
			line_name TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS routes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			origin_seg TEXT NOT NULL,
			dest_seg TEXT NOT NULL,
			path_segs TEXT NOT NULL,
			switches TEXT NOT NULL,
			release TEXT NOT NULL,
			state TEXT NOT NULL DEFAULT 'candidate',
			version_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS versions (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			state TEXT NOT NULL,
			segment_ids TEXT NOT NULL,
			switch_ids TEXT NOT NULL,
			route_ids TEXT NOT NULL,
			exception_ids TEXT NOT NULL,
			conflict_count INTEGER NOT NULL DEFAULT 0,
			last_validated_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS conflicts (
			id TEXT PRIMARY KEY,
			version_id TEXT NOT NULL,
			kind TEXT NOT NULL,
			state TEXT NOT NULL,
			route_a TEXT NOT NULL,
			route_b TEXT NOT NULL DEFAULT '',
			object_id TEXT NOT NULL DEFAULT '',
			detail TEXT NOT NULL DEFAULT '',
			steps TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS exceptions (
			id TEXT PRIMARY KEY,
			version_id TEXT NOT NULL,
			conflict_id TEXT NOT NULL,
			state TEXT NOT NULL,
			reason TEXT NOT NULL,
			owner TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			id TEXT PRIMARY KEY,
			version_id TEXT NOT NULL,
			name TEXT NOT NULL,
			state TEXT NOT NULL,
			topology_hash TEXT NOT NULL,
			conflict_total INTEGER NOT NULL DEFAULT 0,
			exception_count INTEGER NOT NULL DEFAULT 0,
			published_at TEXT,
			superseded_by TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS version_errors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version_id TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
	}
	for _, s := range stmts {
		if _, err := d.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
