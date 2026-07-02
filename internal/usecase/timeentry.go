// Package usecase はアプリケーションのビジネスルールを担う。
// domain のポート(インターフェース)にのみ依存し、DBやHTTPは知らない。
package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/etakahashi78/effort-tracker/internal/domain"
)

// spentOnLayout は spent_on(作業日)の受け入れフォーマット。
const spentOnLayout = "2006-01-02"

// TimeEntryInput は TimeEntry 作成・更新の入力。
// SpentOn は "YYYY-MM-DD" 形式の文字列で受け取り、検証と変換はユースケースが行う。
type TimeEntryInput struct {
	TaskID  int64
	UserID  int64
	Minutes int
	Note    string
	SpentOn string
}

// TimeEntryUsecase は TimeEntry に関するユースケースを実装する。
type TimeEntryUsecase struct {
	repo domain.TimeEntryRepository
}

// NewTimeEntryUsecase は TimeEntryUsecase を生成する。
func NewTimeEntryUsecase(repo domain.TimeEntryRepository) *TimeEntryUsecase {
	return &TimeEntryUsecase{repo: repo}
}

// Create は入力を検証し、新しい工数エントリを作成する。
func (u *TimeEntryUsecase) Create(ctx context.Context, in TimeEntryInput) (*domain.TimeEntry, error) {
	e, err := buildTimeEntry(0, in)
	if err != nil {
		return nil, err
	}
	return u.repo.Create(ctx, e)
}

// List は全工数エントリを返す。
func (u *TimeEntryUsecase) List(ctx context.Context) ([]domain.TimeEntry, error) {
	return u.repo.List(ctx)
}

// Get は ID で工数エントリを取得する。
func (u *TimeEntryUsecase) Get(ctx context.Context, id int64) (*domain.TimeEntry, error) {
	return u.repo.Get(ctx, id)
}

// Update は入力を検証し、既存の工数エントリを更新する。
func (u *TimeEntryUsecase) Update(ctx context.Context, id int64, in TimeEntryInput) (*domain.TimeEntry, error) {
	e, err := buildTimeEntry(id, in)
	if err != nil {
		return nil, err
	}
	return u.repo.Update(ctx, e)
}

// Delete は ID で工数エントリを削除する。
func (u *TimeEntryUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}

// buildTimeEntry は入力を検証し、ドメインエンティティへ変換する(業務ルール)。
func buildTimeEntry(id int64, in TimeEntryInput) (*domain.TimeEntry, error) {
	if in.TaskID == 0 {
		return nil, fmt.Errorf("%w: task_id is required", domain.ErrInvalidInput)
	}
	if in.UserID == 0 {
		return nil, fmt.Errorf("%w: user_id is required", domain.ErrInvalidInput)
	}
	if in.Minutes <= 0 {
		return nil, fmt.Errorf("%w: minutes must be positive", domain.ErrInvalidInput)
	}
	spent := strings.TrimSpace(in.SpentOn)
	if spent == "" {
		return nil, fmt.Errorf("%w: spent_on is required", domain.ErrInvalidInput)
	}
	spentOn, err := time.Parse(spentOnLayout, spent)
	if err != nil {
		return nil, fmt.Errorf("%w: spent_on must be a date (YYYY-MM-DD)", domain.ErrInvalidInput)
	}
	return &domain.TimeEntry{
		ID:      id,
		TaskID:  in.TaskID,
		UserID:  in.UserID,
		Minutes: in.Minutes,
		Note:    in.Note,
		SpentOn: spentOn,
	}, nil
}
