// Package persistence は domain のリポジトリ契約をMySQLで実装する(インターフェースアダプタ層)。
package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/etakahashi78/effort-tracker/internal/domain"
)

// ProjectRepository は domain.ProjectRepository のMySQL実装。
type ProjectRepository struct {
	db *sql.DB
}

// NewProjectRepository は ProjectRepository を生成する。
// 戻り値は具象型だが domain.ProjectRepository を満たす。
func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// 実装漏れをコンパイル時に検出する。
var _ domain.ProjectRepository = (*ProjectRepository)(nil)

// Create は新しいプロジェクトを挿入し、採番されたレコードを返す。
// MySQLは RETURNING 非対応のため LastInsertId 取得後に再取得する。
func (r *ProjectRepository) Create(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	const q = `INSERT INTO projects (name, description, status) VALUES (?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, p.Name, p.Description, p.Status)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

// Get は ID でプロジェクトを取得する。存在しない場合は domain.ErrNotFound。
func (r *ProjectRepository) Get(ctx context.Context, id int64) (*domain.Project, error) {
	const q = `
		SELECT id, name, description, status, created_at, updated_at
		FROM projects WHERE id = ?`
	p, err := scanProject(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return p, err
}

// List は全プロジェクトを作成日時の降順で返す。
func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	const q = `
		SELECT id, name, description, status, created_at, updated_at
		FROM projects ORDER BY created_at DESC, id DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Update は既存プロジェクトを更新する。存在しない場合は domain.ErrNotFound。
// 更新後に再取得して返す(RETURNING 非対応・更新有無に依存しない存在判定のため)。
func (r *ProjectRepository) Update(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	const q = `UPDATE projects SET name = ?, description = ?, status = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, q, p.Name, p.Description, p.Status, p.ID); err != nil {
		return nil, err
	}
	return r.Get(ctx, p.ID)
}

// Delete は ID でプロジェクトを削除する。存在しない場合は domain.ErrNotFound。
func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
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

func scanProject(row *sql.Row) (*domain.Project, error) {
	var p domain.Project
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}
