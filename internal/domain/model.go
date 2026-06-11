// Package domain は最内層のエンティティ・リポジトリ契約・ドメインエラーを定義する。
// 他のどの層にも依存しない(依存性のルール)。
package domain

import "time"

// User はツールを利用するユーザーを表す。
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Project は工数を計上する対象のプロジェクトを表す。
type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // active | archived
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Task はプロジェクト配下の作業単位を表す。
type Task struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // todo | in_progress | done
	AssigneeID  *int64    `json:"assignee_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TimeEntry はタスクに対して計上された工数(作業時間)を表す。
type TimeEntry struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	Minutes   int       `json:"minutes"` // 作業時間(分)
	Note      string    `json:"note"`
	SpentOn   time.Time `json:"spent_on"` // 作業日
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
