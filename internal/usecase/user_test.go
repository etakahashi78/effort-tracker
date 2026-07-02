package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/etakahashi78/effort-tracker/internal/domain"
	"github.com/etakahashi78/effort-tracker/internal/usecase"
)

// mockUserRepo は domain.UserRepository の手書きモック。
type mockUserRepo struct {
	createFn func(ctx context.Context, u *domain.User) (*domain.User, error)
	getFn    func(ctx context.Context, id int64) (*domain.User, error)
	listFn   func(ctx context.Context) ([]domain.User, error)
	updateFn func(ctx context.Context, u *domain.User) (*domain.User, error)
	deleteFn func(ctx context.Context, id int64) error

	createCalls int
	updateCalls int
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	m.createCalls++
	return m.createFn(ctx, u)
}
func (m *mockUserRepo) Get(ctx context.Context, id int64) (*domain.User, error) {
	return m.getFn(ctx, id)
}
func (m *mockUserRepo) List(ctx context.Context) ([]domain.User, error) {
	return m.listFn(ctx)
}
func (m *mockUserRepo) Update(ctx context.Context, u *domain.User) (*domain.User, error) {
	m.updateCalls++
	return m.updateFn(ctx, u)
}
func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

func TestUserUsecase_Create(t *testing.T) {
	tests := []struct {
		name        string
		input       usecase.UserInput
		setupMock   func(m *mockUserRepo, passed **domain.User)
		wantErr     error
		checkResult func(t *testing.T, got *domain.User, passed *domain.User, createCalls int)
	}{
		{
			name: "正常系: 名前とメールをトリムして repo に渡す",
			input: usecase.UserInput{
				Name:  "  太郎  ",
				Email: "  taro@example.com  ",
			},
			setupMock: func(m *mockUserRepo, passed **domain.User) {
				m.createFn = func(_ context.Context, u *domain.User) (*domain.User, error) {
					*passed = u
					u.ID = 1
					return u, nil
				}
			},
			checkResult: func(t *testing.T, got *domain.User, passed *domain.User, _ int) {
				if passed.Name != "太郎" {
					t.Errorf("name should be trimmed: got %q", passed.Name)
				}
				if passed.Email != "taro@example.com" {
					t.Errorf("email should be trimmed: got %q", passed.Email)
				}
				if got.ID != 1 {
					t.Errorf("want id 1, got %d", got.ID)
				}
			},
		},
		{
			name:  "異常系: 名前が空なら ErrInvalidInput で repo を呼ばない",
			input: usecase.UserInput{Name: "   ", Email: "taro@example.com"},
			setupMock: func(m *mockUserRepo, _ **domain.User) {
				m.createFn = func(_ context.Context, u *domain.User) (*domain.User, error) { return u, nil }
			},
			wantErr: domain.ErrInvalidInput,
			checkResult: func(t *testing.T, _ *domain.User, _ *domain.User, createCalls int) {
				if createCalls != 0 {
					t.Errorf("repo.Create should not be called, calls=%d", createCalls)
				}
			},
		},
		{
			name:  "異常系: メールが空なら ErrInvalidInput",
			input: usecase.UserInput{Name: "太郎", Email: ""},
			setupMock: func(m *mockUserRepo, _ **domain.User) {
				m.createFn = func(_ context.Context, u *domain.User) (*domain.User, error) { return u, nil }
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:  "異常系: 不正なメール形式なら ErrInvalidInput (1)",
			input: usecase.UserInput{Name: "太郎", Email: "taroexample.com"},
			setupMock: func(m *mockUserRepo, _ **domain.User) {
				m.createFn = func(_ context.Context, u *domain.User) (*domain.User, error) { return u, nil }
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:  "異常系: 不正なメール形式なら ErrInvalidInput (2)",
			input: usecase.UserInput{Name: "太郎", Email: "taro@"},
			setupMock: func(m *mockUserRepo, _ **domain.User) {
				m.createFn = func(_ context.Context, u *domain.User) (*domain.User, error) { return u, nil }
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:  "異常系: 不正なメール形式なら ErrInvalidInput (3)",
			input: usecase.UserInput{Name: "太郎", Email: "taro@examplecom"},
			setupMock: func(m *mockUserRepo, _ **domain.User) {
				m.createFn = func(_ context.Context, u *domain.User) (*domain.User, error) { return u, nil }
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:  "異常系: repo のエラーを伝播する",
			input: usecase.UserInput{Name: "太郎", Email: "taro@example.com"},
			setupMock: func(m *mockUserRepo, _ **domain.User) {
				m.createFn = func(_ context.Context, _ *domain.User) (*domain.User, error) {
					return nil, errors.New("db error")
				}
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var passed *domain.User
			repo := &mockUserRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo, &passed)
			}
			uc := usecase.NewUserUsecase(repo)

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

func TestUserUsecase_Update(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		input       usecase.UserInput
		setupMock   func(m *mockUserRepo, passed **domain.User)
		wantErr     error
		checkResult func(t *testing.T, got *domain.User, passed *domain.User, updateCalls int)
	}{
		{
			name:  "正常系: 指定 id を保持して repo に渡す",
			id:    7,
			input: usecase.UserInput{Name: "次郎", Email: "jiro@example.com"},
			setupMock: func(m *mockUserRepo, passed **domain.User) {
				m.updateFn = func(_ context.Context, u *domain.User) (*domain.User, error) {
					*passed = u
					return u, nil
				}
			},
			checkResult: func(t *testing.T, _ *domain.User, passed *domain.User, _ int) {
				if passed.ID != 7 {
					t.Errorf("want id 7, got %d", passed.ID)
				}
				if passed.Name != "次郎" || passed.Email != "jiro@example.com" {
					t.Errorf("unexpected entity: %+v", passed)
				}
			},
		},
		{
			name:  "異常系: バリデーション失敗時は repo を呼ばない",
			id:    1,
			input: usecase.UserInput{Name: ""},
			setupMock: func(m *mockUserRepo, _ **domain.User) {
				m.updateFn = func(_ context.Context, u *domain.User) (*domain.User, error) { return u, nil }
			},
			wantErr: domain.ErrInvalidInput,
			checkResult: func(t *testing.T, _ *domain.User, _ *domain.User, updateCalls int) {
				if updateCalls != 0 {
					t.Errorf("repo.Update should not be called, calls=%d", updateCalls)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var passed *domain.User
			repo := &mockUserRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo, &passed)
			}
			uc := usecase.NewUserUsecase(repo)

			got, err := uc.Update(context.Background(), tt.id, tt.input)

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
				tt.checkResult(t, got, passed, repo.updateCalls)
			}
		})
	}
}

func TestUserUsecase_GetListDelete(t *testing.T) {
	repo := &mockUserRepo{
		getFn: func(_ context.Context, id int64) (*domain.User, error) {
			if id == 999 {
				return nil, domain.ErrNotFound
			}
			return &domain.User{ID: id, Name: "A"}, nil
		},
		listFn: func(_ context.Context) ([]domain.User, error) {
			return []domain.User{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}}, nil
		},
		deleteFn: func(_ context.Context, id int64) error {
			if id == 999 {
				return domain.ErrNotFound
			}
			return nil
		},
	}
	uc := usecase.NewUserUsecase(repo)

	// Get
	got, err := uc.Get(context.Background(), 1)
	if err != nil || got.ID != 1 {
		t.Fatalf("Get failed: got=%v, err=%v", got, err)
	}
	_, err = uc.Get(context.Background(), 999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Get 999 should return ErrNotFound, got %v", err)
	}

	// List
	list, err := uc.List(context.Background())
	if err != nil || len(list) != 2 {
		t.Fatalf("List failed: list=%v, err=%v", list, err)
	}

	// Delete
	err = uc.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	err = uc.Delete(context.Background(), 999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Delete 999 should return ErrNotFound, got %v", err)
	}
}
