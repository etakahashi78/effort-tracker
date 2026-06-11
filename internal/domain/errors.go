package domain

import "errors"

var (
	// ErrNotFound はエンティティが存在しない場合に返される。
	ErrNotFound = errors.New("record not found")

	// ErrInvalidInput は入力が業務ルールを満たさない場合に返される。
	// 詳細メッセージは fmt.Errorf("%w: ...", domain.ErrInvalidInput) でラップする。
	ErrInvalidInput = errors.New("invalid input")
)
