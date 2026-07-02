// Package persistence は domain のリポジトリ契約をMySQLで実装する(インターフェースアダプタ層)。
package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/etakahashi78/effort-tracker/internal/domain"
)

// TimeEntryRepository は domain.TimeEntryRepository のMySQL実装。
type TimeEntryRepository struct {
	db *sql.DB
}

// NewTimeEntryRepository は TimeEntryRepository を生成する。
// 戻り値は具象型だが domain.TimeEntryRepository を満たす。
func NewTimeEntryRepository(db *sql.DB) *TimeEntryRepository {
	return &TimeEntryRepository{db: db}
}

// 実装漏れをコンパイル時に検出する。
var _ domain.TimeEntryRepository = (*TimeEntryRepository)(nil)

// Create は新しい工数エントリを挿入し、採番されたレコードを返す。
// MySQLは RETURNING 非対応のため LastInsertId 取得後に再取得する。
func (r *TimeEntryRepository) Create(ctx context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error) {
	const q = `INSERT INTO time_entries (task_id, user_id, minutes, note, spent_on) VALUES (?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q, e.TaskID, e.UserID, e.Minutes, e.Note, e.SpentOn)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

// Get は ID で工数エントリを取得する。存在しない場合は domain.ErrNotFound。
func (r *TimeEntryRepository) Get(ctx context.Context, id int64) (*domain.TimeEntry, error) {
	const q = `
		SELECT id, task_id, user_id, minutes, note, spent_on, created_at, updated_at
		FROM time_entries WHERE id = ?`
	e, err := scanTimeEntry(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return e, err
}

// List は全工数エントリを作業日の降順で返す。
func (r *TimeEntryRepository) List(ctx context.Context) ([]domain.TimeEntry, error) {
	const q = `
		SELECT id, task_id, user_id, minutes, note, spent_on, created_at, updated_at
		FROM time_entries ORDER BY spent_on DESC, id DESC`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.TimeEntry
	for rows.Next() {
		var e domain.TimeEntry
		if err := rows.Scan(&e.ID, &e.TaskID, &e.UserID, &e.Minutes, &e.Note, &e.SpentOn, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// Update は既存の工数エントリを更新する。存在しない場合は domain.ErrNotFound。
// 更新後に再取得して返す(RETURNING 非対応・更新有無に依存しない存在判定のため)。
func (r *TimeEntryRepository) Update(ctx context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error) {
	const q = `UPDATE time_entries SET task_id = ?, user_id = ?, minutes = ?, note = ?, spent_on = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, q, e.TaskID, e.UserID, e.Minutes, e.Note, e.SpentOn, e.ID); err != nil {
		return nil, err
	}
	return r.Get(ctx, e.ID)
}

// Delete は ID で工数エントリを削除する。存在しない場合は domain.ErrNotFound。
func (r *TimeEntryRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM time_entries WHERE id = ?`, id)
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

func scanTimeEntry(row *sql.Row) (*domain.TimeEntry, error) {
	var e domain.TimeEntry
	if err := row.Scan(&e.ID, &e.TaskID, &e.UserID, &e.Minutes, &e.Note, &e.SpentOn, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}
