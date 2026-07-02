// Package usecase はアプリケーションのビジネスルールを担う。
// domain のポート(インターフェース)にのみ依存し、DBやHTTPは知らない。
package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/etakahashi78/effort-tracker/internal/domain"
)

// UserInput は User 作成・更新の入力。
type UserInput struct {
	Name  string
	Email string
}

// UserUsecase は User に関するユースケースを実装する。
type UserUsecase struct {
	repo domain.UserRepository
}

// NewUserUsecase は UserUsecase を生成する。
func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// Create は入力を検証し、新しいユーザーを作成する。
func (u *UserUsecase) Create(ctx context.Context, in UserInput) (*domain.User, error) {
	usr, err := buildUser(0, in)
	if err != nil {
		return nil, err
	}
	return u.repo.Create(ctx, usr)
}

// List は全ユーザーを返す。
func (u *UserUsecase) List(ctx context.Context) ([]domain.User, error) {
	return u.repo.List(ctx)
}

// Get は ID でユーザーを取得する。
func (u *UserUsecase) Get(ctx context.Context, id int64) (*domain.User, error) {
	return u.repo.Get(ctx, id)
}

// Update は入力を検証し、既存ユーザーを更新する。
func (u *UserUsecase) Update(ctx context.Context, id int64, in UserInput) (*domain.User, error) {
	usr, err := buildUser(id, in)
	if err != nil {
		return nil, err
	}
	return u.repo.Update(ctx, usr)
}

// Delete は ID でユーザーを削除する。
func (u *UserUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}

// buildUser は入力を検証し、ドメインエンティティへ変換する(業務ルール)。
func buildUser(id int64, in UserInput) (*domain.User, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}

	email := strings.TrimSpace(in.Email)
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", domain.ErrInvalidInput)
	}

	// 簡易的なメールアドレス形式バリデーション:
	// @ が含まれ、かつ @ の後に . が含まれていることを確認する
	atIdx := strings.Index(email, "@")
	if atIdx == -1 || atIdx == 0 || atIdx == len(email)-1 || !strings.Contains(email[atIdx+1:], ".") {
		return nil, fmt.Errorf("%w: invalid email format", domain.ErrInvalidInput)
	}

	return &domain.User{
		ID:    id,
		Name:  name,
		Email: email,
	}, nil
}
