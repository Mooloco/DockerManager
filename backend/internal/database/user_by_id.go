package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// GetUserByID 按 ID 查询用户
func (d *DB) GetUserByID(id int64) (*User, error) {
	u := &User{}
	err := d.sql.QueryRow(
		"SELECT id, username, password_hash, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}
