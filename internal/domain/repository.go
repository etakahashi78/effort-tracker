package domain

import "context"

// ProjectRepository は Project 永続化のポート(契約)。
// 実装は外側の adapter/persistence 層が提供する(依存性逆転)。
type ProjectRepository interface {
	Create(ctx context.Context, p *Project) (*Project, error)
	Get(ctx context.Context, id int64) (*Project, error)
	List(ctx context.Context) ([]Project, error)
	Update(ctx context.Context, p *Project) (*Project, error)
	Delete(ctx context.Context, id int64) error
}

// TaskRepository は Task 永続化のポート(契約)。
// 実装は外側の adapter/persistence 層が提供する(依存性逆転)。
type TaskRepository interface {
	Create(ctx context.Context, t *Task) (*Task, error)
	Get(ctx context.Context, id int64) (*Task, error)
	List(ctx context.Context) ([]Task, error)
	Update(ctx context.Context, t *Task) (*Task, error)
	Delete(ctx context.Context, id int64) error
}

// TimeEntryRepository は TimeEntry 永続化のポート(契約)。
// 実装は外側の adapter/persistence 層が提供する(依存性逆転)。
type TimeEntryRepository interface {
	Create(ctx context.Context, e *TimeEntry) (*TimeEntry, error)
	Get(ctx context.Context, id int64) (*TimeEntry, error)
	List(ctx context.Context) ([]TimeEntry, error)
	Update(ctx context.Context, e *TimeEntry) (*TimeEntry, error)
	Delete(ctx context.Context, id int64) error
}

// UserRepository は User 永続化のポート(契約)。
// 実装は外側の adapter/persistence 層が提供する(依存性逆転)。
type UserRepository interface {
	Create(ctx context.Context, u *User) (*User, error)
	Get(ctx context.Context, id int64) (*User, error)
	List(ctx context.Context) ([]User, error)
	Update(ctx context.Context, u *User) (*User, error)
	Delete(ctx context.Context, id int64) error
}
