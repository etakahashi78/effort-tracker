// Package persistence は domain のリポジトリ契約をMySQLで実装する(インターフェースアダプタ層)。
package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/etakahashi78/effort-tracker/internal/domain"
)

// TaskRepository は domain.TaskRepository のMySQL実装。
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository は TaskRepository を生成する。
// 戻り値は具象型だが domain.TaskRepository を満たす。
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// 実装漏れをコンパイル時に検出する。
var _ domain.TaskRepository = (*TaskRepository)(nil)

// Create は新しいタスクを挿入し、採番されたレコードを返す。
// MySQLは RETURNING 非対応のため LastInsertId 取得後に再取得する。
func (r *TaskRepository) Create(ctx context.Context, t *domain.Task) (*domain.Task, error) {
	const q = `INSERT INTO tasks (project_id, title, description, status, assignee_id) VALUES (?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, t.ProjectID, t.Title, t.Description, t.Status, t.AssigneeID)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

// Get は ID でタスクを取得する。存在しない場合は domain.ErrNotFound。
func (r *TaskRepository) Get(ctx context.Context, id int64) (*domain.Task, error) {
	const q = `
		SELECT id, project_id, title, description, status, assignee_id, created_at, updated_at
		FROM tasks WHERE id = ?`
	t, err := scanTask(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return t, err
}

// List は全タスクを作成日時の降順で返す。
func (r *TaskRepository) List(ctx context.Context) ([]domain.Task, error) {
	const q = `
		SELECT id, project_id, title, description, status, assignee_id, created_at, updated_at
		FROM tasks ORDER BY created_at DESC, id DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// Update は既存タスクを更新する。存在しない場合は domain.ErrNotFound。
// 更新後に再取得して返す(RETURNING 非対応・更新有無に依存しない存在判定のため)。
func (r *TaskRepository) Update(ctx context.Context, t *domain.Task) (*domain.Task, error) {
	const q = `UPDATE tasks SET project_id = ?, title = ?, description = ?, status = ?, assignee_id = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, q, t.ProjectID, t.Title, t.Description, t.Status, t.AssigneeID, t.ID); err != nil {
		return nil, err
	}
	return r.Get(ctx, t.ID)
}

// Delete は ID でタスクを削除する。存在しない場合は domain.ErrNotFound。
func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
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

func scanTask(row *sql.Row) (*domain.Task, error) {
	var t domain.Task
	if err := row.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	return &t, nil
}
