package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/etakahashi78/effort-tracker/internal/domain"
	"github.com/etakahashi78/effort-tracker/internal/usecase"
)

// mockProjectRepo は domain.ProjectRepository の手書きモック。
// 各メソッドの挙動を関数フィールドで差し替えられる。
type mockProjectRepo struct {
	createFn func(ctx context.Context, p *domain.Project) (*domain.Project, error)
	getFn    func(ctx context.Context, id int64) (*domain.Project, error)
	listFn   func(ctx context.Context) ([]domain.Project, error)
	updateFn func(ctx context.Context, p *domain.Project) (*domain.Project, error)
	deleteFn func(ctx context.Context, id int64) error

	createCalls int
	updateCalls int
}

func (m *mockProjectRepo) Create(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	m.createCalls++
	return m.createFn(ctx, p)
}
func (m *mockProjectRepo) Get(ctx context.Context, id int64) (*domain.Project, error) {
	return m.getFn(ctx, id)
}
func (m *mockProjectRepo) List(ctx context.Context) ([]domain.Project, error) {
	return m.listFn(ctx)
}
func (m *mockProjectRepo) Update(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	m.updateCalls++
	return m.updateFn(ctx, p)
}
func (m *mockProjectRepo) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

func TestProjectUsecase_Create(t *testing.T) {
	tests := []struct {
		name        string
		input       usecase.ProjectInput
		setupMock   func(m *mockProjectRepo, passed **domain.Project)
		wantErr     error
		checkResult func(t *testing.T, got *domain.Project, passed *domain.Project, createCalls int)
	}{
		{
			name: "正常系: 名前をトリムし status 既定値を補完して repo に渡す",
			input: usecase.ProjectInput{
				Name:        "  社内ツール  ",
				Description: "工数管理",
			},
			setupMock: func(m *mockProjectRepo, passed **domain.Project) {
				m.createFn = func(_ context.Context, p *domain.Project) (*domain.Project, error) {
					*passed = p
					p.ID = 1
					return p, nil
				}
			},
			checkResult: func(t *testing.T, got *domain.Project, passed *domain.Project, _ int) {
				if passed.Name != "社内ツール" {
					t.Errorf("name should be trimmed: got %q", passed.Name)
				}
				if passed.Status != "active" {
					t.Errorf("status should default to active: got %q", passed.Status)
				}
				if got.ID != 1 {
					t.Errorf("want id 1, got %d", got.ID)
				}
			},
		},
		{
			name:  "異常系: 名前が空白のみなら ErrInvalidInput で repo を呼ばない",
			input: usecase.ProjectInput{Name: "   "},
			setupMock: func(m *mockProjectRepo, _ **domain.Project) {
				m.createFn = func(_ context.Context, p *domain.Project) (*domain.Project, error) {
					return p, nil
				}
			},
			wantErr: domain.ErrInvalidInput,
			checkResult: func(t *testing.T, _ *domain.Project, _ *domain.Project, createCalls int) {
				if createCalls != 0 {
					t.Errorf("repo.Create should not be called, calls=%d", createCalls)
				}
			},
		},
		{
			name:  "異常系: 不正な status は ErrInvalidInput",
			input: usecase.ProjectInput{Name: "x", Status: "bogus"},
			setupMock: func(m *mockProjectRepo, _ **domain.Project) {
				m.createFn = func(_ context.Context, p *domain.Project) (*domain.Project, error) { return p, nil }
			},
			wantErr: domain.ErrInvalidInput,
			checkResult: func(t *testing.T, _ *domain.Project, _ *domain.Project, createCalls int) {
				if createCalls != 0 {
					t.Errorf("repo.Create should not be called, calls=%d", createCalls)
				}
			},
		},
		{
			name:  "異常系: repo のエラーをそのまま伝播する",
			input: usecase.ProjectInput{Name: "x"},
			setupMock: func(m *mockProjectRepo, _ **domain.Project) {
				m.createFn = func(_ context.Context, _ *domain.Project) (*domain.Project, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var passed *domain.Project
			repo := &mockProjectRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo, &passed)
			}
			uc := usecase.NewProjectUsecase(repo)

			got, err := uc.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil || (tt.wantErr.Error() != err.Error() && !errors.Is(err, tt.wantErr)) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if tt.checkResult != nil {
				tt.checkResult(t, got, passed, repo.createCalls)
			}
		})
	}
}

func TestProjectUsecase_Update(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		input       usecase.ProjectInput
		setupMock   func(m *mockProjectRepo, passed **domain.Project)
		wantErr     error
		checkResult func(t *testing.T, got *domain.Project, passed *domain.Project, updateCalls int)
	}{
		{
			name:  "正常系: 指定 id を保持して repo に渡す",
			id:    7,
			input: usecase.ProjectInput{Name: "x", Status: "archived"},
			setupMock: func(m *mockProjectRepo, passed **domain.Project) {
				m.updateFn = func(_ context.Context, p *domain.Project) (*domain.Project, error) {
					*passed = p
					return p, nil
				}
			},
			checkResult: func(t *testing.T, _ *domain.Project, passed *domain.Project, _ int) {
				if passed.ID != 7 {
					t.Errorf("want id 7, got %d", passed.ID)
				}
				if passed.Status != "archived" {
					t.Errorf("want status archived, got %q", passed.Status)
				}
			},
		},
		{
			name:  "異常系: バリデーション失敗時は repo を呼ばない",
			id:    1,
			input: usecase.ProjectInput{Name: ""},
			setupMock: func(m *mockProjectRepo, _ **domain.Project) {
				m.updateFn = func(_ context.Context, p *domain.Project) (*domain.Project, error) { return p, nil }
			},
			wantErr: domain.ErrInvalidInput,
			checkResult: func(t *testing.T, _ *domain.Project, _ *domain.Project, updateCalls int) {
				if updateCalls != 0 {
					t.Errorf("repo.Update should not be called, calls=%d", updateCalls)
				}
			},
		},
		{
			name:  "異常系: repo の ErrNotFound を伝播する",
			id:    1,
			input: usecase.ProjectInput{Name: "x"},
			setupMock: func(m *mockProjectRepo, _ **domain.Project) {
				m.updateFn = func(_ context.Context, _ *domain.Project) (*domain.Project, error) {
					return nil, domain.ErrNotFound
				}
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var passed *domain.Project
			repo := &mockProjectRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo, &passed)
			}
			uc := usecase.NewProjectUsecase(repo)

			got, err := uc.Update(context.Background(), tt.id, tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if tt.checkResult != nil {
				tt.checkResult(t, got, passed, repo.updateCalls)
			}
		})
	}
}

func TestProjectUsecase_Get(t *testing.T) {
	want := &domain.Project{ID: 3, Name: "x"}
	tests := []struct {
		name      string
		id        int64
		setupMock func(m *mockProjectRepo)
		want      *domain.Project
		wantErr   error
	}{
		{
			name: "repo の結果を返す",
			id:   3,
			setupMock: func(m *mockProjectRepo) {
				m.getFn = func(_ context.Context, id int64) (*domain.Project, error) {
					return want, nil
				}
			},
			want: want,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}
			uc := usecase.NewProjectUsecase(repo)
			got, err := uc.Get(context.Background(), tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want error %v, got %v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectUsecase_List(t *testing.T) {
	want := []domain.Project{{ID: 1}, {ID: 2}}
	tests := []struct {
		name      string
		setupMock func(m *mockProjectRepo)
		want      []domain.Project
		wantErr   error
	}{
		{
			name: "repo の結果を返す",
			setupMock: func(m *mockProjectRepo) {
				m.listFn = func(_ context.Context) ([]domain.Project, error) { return want, nil }
			},
			want: want,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}
			uc := usecase.NewProjectUsecase(repo)
			got, err := uc.List(context.Background())
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want error %v, got %v", tt.wantErr, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got length %d, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestProjectUsecase_Delete(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		setupMock func(m *mockProjectRepo)
		wantErr   error
	}{
		{
			name: "repo のエラーを伝播する",
			id:   99,
			setupMock: func(m *mockProjectRepo) {
				m.deleteFn = func(_ context.Context, _ int64) error { return domain.ErrNotFound }
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}
			uc := usecase.NewProjectUsecase(repo)
			err := uc.Delete(context.Background(), tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
