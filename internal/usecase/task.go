// Package usecase はアプリケーションのビジネスルールを担う。
// domain のポート(インターフェース)にのみ依存し、DBやHTTPは知らない。
package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/etakahashi78/effort-tracker/internal/domain"
)

// TaskInput は Task 作成・更新の入力。
type TaskInput struct {
	ProjectID   int64
	Title       string
	Description string
	Status      string
	AssigneeID  *int64
}

// TaskUsecase は Task に関するユースケースを実装する。
type TaskUsecase struct {
	repo domain.TaskRepository
}

// NewTaskUsecase は TaskUsecase を生成する。
func NewTaskUsecase(repo domain.TaskRepository) *TaskUsecase {
	return &TaskUsecase{repo: repo}
}

// Create は入力を検証し、新しいタスクを作成する。
func (u *TaskUsecase) Create(ctx context.Context, in TaskInput) (*domain.Task, error) {
	t, err := buildTask(0, in)
	if err != nil {
		return nil, err
	}
	return u.repo.Create(ctx, t)
}

// List は全タスクを返す。
func (u *TaskUsecase) List(ctx context.Context) ([]domain.Task, error) {
	return u.repo.List(ctx)
}

// Get は ID でタスクを取得する。
func (u *TaskUsecase) Get(ctx context.Context, id int64) (*domain.Task, error) {
	return u.repo.Get(ctx, id)
}

// Update は入力を検証し、既存タスクを更新する。
func (u *TaskUsecase) Update(ctx context.Context, id int64, in TaskInput) (*domain.Task, error) {
	t, err := buildTask(id, in)
	if err != nil {
		return nil, err
	}
	return u.repo.Update(ctx, t)
}

// Delete は ID でタスクを削除する。
func (u *TaskUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}

// buildTask は入力を検証し、ドメインエンティティへ変換する(業務ルール)。
func buildTask(id int64, in TaskInput) (*domain.Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}
	if in.ProjectID == 0 {
		return nil, fmt.Errorf("%w: project_id is required", domain.ErrInvalidInput)
	}
	status := in.Status
	if status == "" {
		status = "todo"
	}
	if status != "todo" && status != "in_progress" && status != "done" {
		return nil, fmt.Errorf("%w: status must be 'todo', 'in_progress' or 'done'", domain.ErrInvalidInput)
	}
	return &domain.Task{
		ID:          id,
		ProjectID:   in.ProjectID,
		Title:       title,
		Description: in.Description,
		Status:      status,
		AssigneeID:  in.AssigneeID,
	}, nil
}
