package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB 包装 SQLite 连接
type DB struct {
	sql *sql.DB
}

// Open 打开(必要时创建)SQLite 数据库并初始化 schema
func Open(path string) (*DB, error) {
	// 确保数据库文件所在目录存在
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %w", err)
		}
	}
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	// modernc sqlite 默认单连接即可,设置连接池上限避免锁竞争
	sqlDB.SetMaxOpenConns(1)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

// migrate 创建所需的表(轻量迁移,幂等)
func (d *DB) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS sessions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    token         TEXT    NOT NULL UNIQUE,
    user_id       INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    last_seen_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    expires_at    TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_sessions_user  ON sessions(user_id);
`
	if _, err := d.sql.Exec(schema); err != nil {
		return fmt.Errorf("初始化数据库 schema 失败: %w", err)
	}
	slog.Info("数据库 schema 就绪")
	return nil
}

// Close 关闭数据库
func (d *DB) Close() error {
	return d.sql.Close()
}

// SQL 暴露底层连接(供各 store 使用)
func (d *DB) SQL() *sql.DB {
	return d.sql
}
