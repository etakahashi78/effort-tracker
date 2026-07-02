// Package persistence は domain のリポジトリ契約をMySQLで実装する(インターフェースアダプタ層)。
package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/etakahashi78/effort-tracker/internal/domain"
)

// UserRepository は domain.UserRepository のMySQL実装。
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository は UserRepository を生成する。
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// 実装漏れをコンパイル時に検出する。
var _ domain.UserRepository = (*UserRepository)(nil)

// Create は新しいユーザーを挿入し、採番されたレコードを返す。
func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	const q = `INSERT INTO users (name, email) VALUES (?, ?)`
	res, err := r.db.ExecContext(ctx, q, u.Name, u.Email)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

// Get は ID でユーザーを取得する。存在しない場合は domain.ErrNotFound。
func (r *UserRepository) Get(ctx context.Context, id int64) (*domain.User, error) {
	const q = `
		SELECT id, name, email, created_at, updated_at
		FROM users WHERE id = ?`
	u, err := scanUser(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return u, err
}

// List は全ユーザーを作成日時の降順で返す。
func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	const q = `
		SELECT id, name, email, created_at, updated_at
		FROM users ORDER BY created_at DESC, id DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// Update は既存ユーザーを更新する。存在しない場合は domain.ErrNotFound。
func (r *UserRepository) Update(ctx context.Context, u *domain.User) (*domain.User, error) {
	const q = `UPDATE users SET name = ?, email = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, q, u.Name, u.Email, u.ID); err != nil {
		return nil, err
	}
	return r.Get(ctx, u.ID)
}

// Delete は ID でユーザーを削除する。存在しない場合は domain.ErrNotFound。
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
