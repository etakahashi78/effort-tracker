package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/etakahashi78/effort-tracker/internal/domain"
	"github.com/etakahashi78/effort-tracker/internal/usecase"
)

// mockTimeEntryRepo は domain.TimeEntryRepository の手書きモック。
// 各メソッドの挙動を関数フィールドで差し替えられる。
type mockTimeEntryRepo struct {
	createFn func(ctx context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error)
	getFn    func(ctx context.Context, id int64) (*domain.TimeEntry, error)
	listFn   func(ctx context.Context) ([]domain.TimeEntry, error)
	updateFn func(ctx context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error)
	deleteFn func(ctx context.Context, id int64) error

	createCalls int
	updateCalls int
}

func (m *mockTimeEntryRepo) Create(ctx context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error) {
	m.createCalls++
	return m.createFn(ctx, e)
}
func (m *mockTimeEntryRepo) Get(ctx context.Context, id int64) (*domain.TimeEntry, error) {
	return m.getFn(ctx, id)
}
func (m *mockTimeEntryRepo) List(ctx context.Context) ([]domain.TimeEntry, error) {
	return m.listFn(ctx)
}
func (m *mockTimeEntryRepo) Update(ctx context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error) {
	m.updateCalls++
	return m.updateFn(ctx, e)
}
func (m *mockTimeEntryRepo) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

func validTimeEntryInput() usecase.TimeEntryInput {
	return usecase.TimeEntryInput{TaskID: 1, UserID: 2, Minutes: 30, Note: "作業", SpentOn: "2026-07-02"}
}

func TestTimeEntryUsecase_Create(t *testing.T) {
	tests := []struct {
		name        string
		input       usecase.TimeEntryInput
		setupMock   func(m *mockTimeEntryRepo, passed **domain.TimeEntry)
		wantErr     error
		checkResult func(t *testing.T, got *domain.TimeEntry, passed *domain.TimeEntry, createCalls int)
	}{
		{
			name:  "正常系: spent_on を time.Time へ変換し repo に渡す",
			input: validTimeEntryInput(),
			setupMock: func(m *mockTimeEntryRepo, passed **domain.TimeEntry) {
				m.createFn = func(_ context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error) {
					*passed = e
					e.ID = 1
					return e, nil
				}
			},
			checkResult: func(t *testing.T, got *domain.TimeEntry, passed *domain.TimeEntry, _ int) {
				if passed.TaskID != 1 || passed.UserID != 2 || passed.Minutes != 30 {
					t.Errorf("unexpected entity passed to repo: %+v", passed)
				}
				if y, mo, d := passed.SpentOn.Date(); y != 2026 || mo != 7 || d != 2 {
					t.Errorf("spent_on not parsed correctly: %v", passed.SpentOn)
				}
				if got.ID != 1 {
					t.Errorf("want id 1, got %d", got.ID)
				}
			},
		},
		{
			name:  "異常系: task_id 未指定は ErrInvalidInput で repo を呼ばない",
			input: usecase.TimeEntryInput{UserID: 2, Minutes: 30, SpentOn: "2026-07-02"},
			setupMock: func(m *mockTimeEntryRepo, _ **domain.TimeEntry) {
				m.createFn = func(_ context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error) { return e, nil }
			},
			wantErr: domain.ErrInvalidInput,
			checkResult: func(t *testing.T, _ *domain.TimeEntry, _ *domain.TimeEntry, createCalls int) {
				if createCalls != 0 {
					t.Errorf("repo.Create should not be called, calls=%d", createCalls)
				}
			},
		},
		{
			name:    "異常系: user_id 未指定は ErrInvalidInput",
			input:   usecase.TimeEntryInput{TaskID: 1, Minutes: 30, SpentOn: "2026-07-02"},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:    "異常系: minutes が 0 以下は ErrInvalidInput",
			input:   usecase.TimeEntryInput{TaskID: 1, UserID: 2, Minutes: 0, SpentOn: "2026-07-02"},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:    "異常系: spent_on が空白のみは ErrInvalidInput",
			input:   usecase.TimeEntryInput{TaskID: 1, UserID: 2, Minutes: 30, SpentOn: "   "},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:    "異常系: spent_on の形式不正は ErrInvalidInput",
			input:   usecase.TimeEntryInput{TaskID: 1, UserID: 2, Minutes: 30, SpentOn: "2026/07/02"},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:  "異常系: repo のエラーをそのまま伝播する",
			input: validTimeEntryInput(),
			setupMock: func(m *mockTimeEntryRepo, _ **domain.TimeEntry) {
				m.createFn = func(_ context.Context, _ *domain.TimeEntry) (*domain.TimeEntry, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var passed *domain.TimeEntry
			repo := &mockTimeEntryRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo, &passed)
			}
			uc := usecase.NewTimeEntryUsecase(repo)

			got, err := uc.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil || (tt.wantErr.Error() != err.Error() && !errors.Is(err, tt.wantErr)) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, got, passed, repo.createCalls)
			}
		})
	}
}

func TestTimeEntryUsecase_Update(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		input       usecase.TimeEntryInput
		setupMock   func(m *mockTimeEntryRepo, passed **domain.TimeEntry)
		wantErr     error
		checkResult func(t *testing.T, passed *domain.TimeEntry, updateCalls int)
	}{
		{
			name:  "正常系: 指定 id を保持して repo に渡す",
			id:    7,
			input: validTimeEntryInput(),
			setupMock: func(m *mockTimeEntryRepo, passed **domain.TimeEntry) {
				m.updateFn = func(_ context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error) {
					*passed = e
					return e, nil
				}
			},
			checkResult: func(t *testing.T, passed *domain.TimeEntry, _ int) {
				if passed.ID != 7 {
					t.Errorf("want id 7, got %d", passed.ID)
				}
			},
		},
		{
			name:  "異常系: バリデーション失敗時は repo を呼ばない",
			id:    1,
			input: usecase.TimeEntryInput{TaskID: 1, UserID: 2, Minutes: -1, SpentOn: "2026-07-02"},
			setupMock: func(m *mockTimeEntryRepo, _ **domain.TimeEntry) {
				m.updateFn = func(_ context.Context, e *domain.TimeEntry) (*domain.TimeEntry, error) { return e, nil }
			},
			wantErr: domain.ErrInvalidInput,
			checkResult: func(t *testing.T, _ *domain.TimeEntry, updateCalls int) {
				if updateCalls != 0 {
					t.Errorf("repo.Update should not be called, calls=%d", updateCalls)
				}
			},
		},
		{
			name:  "異常系: repo の ErrNotFound を伝播する",
			id:    1,
			input: validTimeEntryInput(),
			setupMock: func(m *mockTimeEntryRepo, _ **domain.TimeEntry) {
				m.updateFn = func(_ context.Context, _ *domain.TimeEntry) (*domain.TimeEntry, error) {
					return nil, domain.ErrNotFound
				}
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var passed *domain.TimeEntry
			repo := &mockTimeEntryRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo, &passed)
			}
			uc := usecase.NewTimeEntryUsecase(repo)

			_, err := uc.Update(context.Background(), tt.id, tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, passed, repo.updateCalls)
			}
		})
	}
}

func TestTimeEntryUsecase_Get(t *testing.T) {
	want := &domain.TimeEntry{ID: 3}
	repo := &mockTimeEntryRepo{
		getFn: func(_ context.Context, _ int64) (*domain.TimeEntry, error) { return want, nil },
	}
	uc := usecase.NewTimeEntryUsecase(repo)
	got, err := uc.Get(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestTimeEntryUsecase_List(t *testing.T) {
	want := []domain.TimeEntry{{ID: 1}, {ID: 2}}
	repo := &mockTimeEntryRepo{
		listFn: func(_ context.Context) ([]domain.TimeEntry, error) { return want, nil },
	}
	uc := usecase.NewTimeEntryUsecase(repo)
	got, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got length %d, want %d", len(got), len(want))
	}
}

func TestTimeEntryUsecase_Delete(t *testing.T) {
	repo := &mockTimeEntryRepo{
		deleteFn: func(_ context.Context, _ int64) error { return domain.ErrNotFound },
	}
	uc := usecase.NewTimeEntryUsecase(repo)
	if err := uc.Delete(context.Background(), 99); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
