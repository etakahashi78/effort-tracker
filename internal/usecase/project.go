// Package usecase はアプリケーションのビジネスルールを担う。
// domain のポート(インターフェース)にのみ依存し、DBやHTTPは知らない。
package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/etakahashi78/effort-tracker/internal/domain"
)

// ProjectInput は Project 作成・更新の入力。
type ProjectInput struct {
	Name        string
	Description string
	Status      string
}

// ProjectUsecase は Project に関するユースケースを実装する。
type ProjectUsecase struct {
	repo domain.ProjectRepository
}

// NewProjectUsecase は ProjectUsecase を生成する。
func NewProjectUsecase(repo domain.ProjectRepository) *ProjectUsecase {
	return &ProjectUsecase{repo: repo}
}

// Create は入力を検証し、新しいプロジェクトを作成する。
func (u *ProjectUsecase) Create(ctx context.Context, in ProjectInput) (*domain.Project, error) {
	p, err := buildProject(0, in)
	if err != nil {
		return nil, err
	}
	return u.repo.Create(ctx, p)
}

// List は全プロジェクトを返す。
func (u *ProjectUsecase) List(ctx context.Context) ([]domain.Project, error) {
	return u.repo.List(ctx)
}

// Get は ID でプロジェクトを取得する。
func (u *ProjectUsecase) Get(ctx context.Context, id int64) (*domain.Project, error) {
	return u.repo.Get(ctx, id)
}

// Update は入力を検証し、既存プロジェクトを更新する。
func (u *ProjectUsecase) Update(ctx context.Context, id int64, in ProjectInput) (*domain.Project, error) {
	p, err := buildProject(id, in)
	if err != nil {
		return nil, err
	}
	return u.repo.Update(ctx, p)
}

// Delete は ID でプロジェクトを削除する。
func (u *ProjectUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}

// buildProject は入力を検証し、ドメインエンティティへ変換する(業務ルール)。
func buildProject(id int64, in ProjectInput) (*domain.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	status := in.Status
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "archived" {
		return nil, fmt.Errorf("%w: status must be 'active' or 'archived'", domain.ErrInvalidInput)
	}
	return &domain.Project{
		ID:          id,
		Name:        name,
		Description: in.Description,
		Status:      status,
	}, nil
}
