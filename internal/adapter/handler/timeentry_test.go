package handler_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/etakahashi78/effort-tracker/internal/adapter/handler"
	"github.com/etakahashi78/effort-tracker/internal/domain"
	"github.com/etakahashi78/effort-tracker/internal/infra/router"
	"github.com/etakahashi78/effort-tracker/internal/usecase"
)

// mockTimeEntryUsecase は handler.TimeEntryUsecase の手書きモック。
type mockTimeEntryUsecase struct {
	createFn func(ctx context.Context, in usecase.TimeEntryInput) (*domain.TimeEntry, error)
	listFn   func(ctx context.Context) ([]domain.TimeEntry, error)
	getFn    func(ctx context.Context, id int64) (*domain.TimeEntry, error)
	updateFn func(ctx context.Context, id int64, in usecase.TimeEntryInput) (*domain.TimeEntry, error)
	deleteFn func(ctx context.Context, id int64) error
}

func (m *mockTimeEntryUsecase) Create(ctx context.Context, in usecase.TimeEntryInput) (*domain.TimeEntry, error) {
	return m.createFn(ctx, in)
}
func (m *mockTimeEntryUsecase) List(ctx context.Context) ([]domain.TimeEntry, error) {
	return m.listFn(ctx)
}
func (m *mockTimeEntryUsecase) Get(ctx context.Context, id int64) (*domain.TimeEntry, error) {
	return m.getFn(ctx, id)
}
func (m *mockTimeEntryUsecase) Update(ctx context.Context, id int64, in usecase.TimeEntryInput) (*domain.TimeEntry, error) {
	return m.updateFn(ctx, id, in)
}
func (m *mockTimeEntryUsecase) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

// newTimeEntryTestServer はモックを注入したハンドラを実ルータに載せた http.Handler を返す。
func newTimeEntryTestServer(uc handler.TimeEntryUsecase) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return router.New(logger, handler.NewProjectHandler(nil), handler.NewTimeEntryHandler(uc))
}

func TestTimeEntryHandler_Create(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		setupMock     func(m *mockTimeEntryUsecase, gotIn *usecase.TimeEntryInput, called *bool)
		wantStatus    int
		checkInput    func(t *testing.T, gotIn usecase.TimeEntryInput, called bool)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 201 と作成結果のJSONを返す",
			body: `{"task_id":1,"user_id":2,"minutes":30,"note":"作業","spent_on":"2026-07-02"}`,
			setupMock: func(m *mockTimeEntryUsecase, gotIn *usecase.TimeEntryInput, _ *bool) {
				m.createFn = func(_ context.Context, in usecase.TimeEntryInput) (*domain.TimeEntry, error) {
					*gotIn = in
					return &domain.TimeEntry{ID: 1, TaskID: in.TaskID, UserID: in.UserID, Minutes: in.Minutes}, nil
				}
			},
			wantStatus: http.StatusCreated,
			checkInput: func(t *testing.T, gotIn usecase.TimeEntryInput, _ bool) {
				if gotIn.TaskID != 1 || gotIn.UserID != 2 || gotIn.Minutes != 30 || gotIn.SpentOn != "2026-07-02" {
					t.Errorf("usecase received unexpected input: %+v", gotIn)
				}
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got domain.TimeEntry
				mustDecode(t, rec, &got)
				if got.ID != 1 {
					t.Errorf("want id 1, got %d", got.ID)
				}
			},
		},
		{
			name: "異常系: 不正なJSONは usecase を呼ばず 400",
			body: `{`,
			setupMock: func(m *mockTimeEntryUsecase, _ *usecase.TimeEntryInput, called *bool) {
				m.createFn = func(_ context.Context, _ usecase.TimeEntryInput) (*domain.TimeEntry, error) {
					*called = true
					return nil, nil
				}
			},
			wantStatus: http.StatusBadRequest,
			checkInput: func(t *testing.T, _ usecase.TimeEntryInput, called bool) {
				if called {
					t.Error("usecase should not be called on bad JSON")
				}
			},
		},
		{
			name: "異常系: usecase の ErrInvalidInput を 400 に変換",
			body: `{}`,
			setupMock: func(m *mockTimeEntryUsecase, _ *usecase.TimeEntryInput, _ *bool) {
				m.createFn = func(_ context.Context, _ usecase.TimeEntryInput) (*domain.TimeEntry, error) {
					return nil, fmt.Errorf("%w: task_id is required", domain.ErrInvalidInput)
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotIn usecase.TimeEntryInput
			called := false
			m := &mockTimeEntryUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m, &gotIn, &called)
			}
			h := newTimeEntryTestServer(m)

			rec := do(t, h, http.MethodPost, "/time-entries", tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d (%s)", tt.wantStatus, rec.Code, rec.Body)
			}
			if tt.checkInput != nil {
				tt.checkInput(t, gotIn, called)
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestTimeEntryHandler_Get(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		setupMock     func(m *mockTimeEntryUsecase, called *bool)
		wantStatus    int
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 200(chi URL パラメータ解決)",
			path: "/time-entries/5",
			setupMock: func(m *mockTimeEntryUsecase, _ *bool) {
				m.getFn = func(_ context.Context, id int64) (*domain.TimeEntry, error) {
					return &domain.TimeEntry{ID: id}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got domain.TimeEntry
				mustDecode(t, rec, &got)
				if got.ID != 5 {
					t.Errorf("chi URL param not resolved: want id 5, got %d", got.ID)
				}
			},
		},
		{
			name: "異常系: ErrNotFound を 404 に変換",
			path: "/time-entries/999",
			setupMock: func(m *mockTimeEntryUsecase, _ *bool) {
				m.getFn = func(_ context.Context, _ int64) (*domain.TimeEntry, error) {
					return nil, domain.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "異常系: 数値でない id は usecase を呼ばず 400",
			path: "/time-entries/abc",
			setupMock: func(m *mockTimeEntryUsecase, called *bool) {
				m.getFn = func(_ context.Context, _ int64) (*domain.TimeEntry, error) {
					*called = true
					return nil, nil
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			m := &mockTimeEntryUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m, &called)
			}
			h := newTimeEntryTestServer(m)

			rec := do(t, h, http.MethodGet, tt.path, "")

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d", tt.wantStatus, rec.Code)
			}
			if tt.name == "異常系: 数値でない id は usecase を呼ばず 400" && called {
				t.Error("usecase should not be called on invalid id")
			}
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestTimeEntryHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func(m *mockTimeEntryUsecase, gotID *int64, gotIn *usecase.TimeEntryInput, called *bool)
		wantStatus int
		checkInput func(t *testing.T, gotID int64, gotIn usecase.TimeEntryInput, called bool)
	}{
		{
			name: "正常系: 200、id とボディを usecase に渡す",
			path: "/time-entries/8",
			body: `{"task_id":1,"user_id":2,"minutes":45,"spent_on":"2026-07-02"}`,
			setupMock: func(m *mockTimeEntryUsecase, gotID *int64, gotIn *usecase.TimeEntryInput, _ *bool) {
				m.updateFn = func(_ context.Context, id int64, in usecase.TimeEntryInput) (*domain.TimeEntry, error) {
					*gotID, *gotIn = id, in
					return &domain.TimeEntry{ID: id, Minutes: in.Minutes}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkInput: func(t *testing.T, gotID int64, gotIn usecase.TimeEntryInput, _ bool) {
				if gotID != 8 {
					t.Errorf("want id 8, got %d", gotID)
				}
				if gotIn.Minutes != 45 {
					t.Errorf("usecase received unexpected input: %+v", gotIn)
				}
			},
		},
		{
			name: "異常系: ErrNotFound を 404 に変換",
			path: "/time-entries/999",
			body: `{"task_id":1,"user_id":2,"minutes":45,"spent_on":"2026-07-02"}`,
			setupMock: func(m *mockTimeEntryUsecase, _ *int64, _ *usecase.TimeEntryInput, _ *bool) {
				m.updateFn = func(_ context.Context, _ int64, _ usecase.TimeEntryInput) (*domain.TimeEntry, error) {
					return nil, domain.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "異常系: 不正なJSONは 400",
			path: "/time-entries/1",
			body: `{bad`,
			setupMock: func(m *mockTimeEntryUsecase, _ *int64, _ *usecase.TimeEntryInput, called *bool) {
				m.updateFn = func(_ context.Context, _ int64, _ usecase.TimeEntryInput) (*domain.TimeEntry, error) {
					*called = true
					return nil, nil
				}
			},
			wantStatus: http.StatusBadRequest,
			checkInput: func(t *testing.T, _ int64, _ usecase.TimeEntryInput, called bool) {
				if called {
					t.Error("usecase should not be called on bad JSON")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID int64
			var gotIn usecase.TimeEntryInput
			called := false
			m := &mockTimeEntryUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m, &gotID, &gotIn, &called)
			}
			h := newTimeEntryTestServer(m)

			rec := do(t, h, http.MethodPut, tt.path, tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d (%s)", tt.wantStatus, rec.Code, rec.Body)
			}
			if tt.checkInput != nil {
				tt.checkInput(t, gotID, gotIn, called)
			}
		})
	}
}

func TestTimeEntryHandler_List(t *testing.T) {
	tests := []struct {
		name    string
		listFn  func(ctx context.Context) ([]domain.TimeEntry, error)
		wantLen int
	}{
		{
			name:    "正常系: 200",
			listFn:  func(_ context.Context) ([]domain.TimeEntry, error) { return []domain.TimeEntry{{ID: 1}, {ID: 2}}, nil },
			wantLen: 2,
		},
		{
			name:    "正常系: nil は空配列 [] として返す",
			listFn:  func(_ context.Context) ([]domain.TimeEntry, error) { return nil, nil },
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTimeEntryTestServer(&mockTimeEntryUsecase{listFn: tt.listFn})

			rec := do(t, h, http.MethodGet, "/time-entries", "")

			if rec.Code != http.StatusOK {
				t.Fatalf("want 200, got %d", rec.Code)
			}
			var got []domain.TimeEntry
			mustDecode(t, rec, &got)
			if len(got) != tt.wantLen {
				t.Errorf("want %d entries, got %d", tt.wantLen, len(got))
			}
			if got == nil {
				t.Error("response should be [] not null")
			}
		})
	}
}

func TestTimeEntryHandler_InvalidID(t *testing.T) {
	// パスIDが数値でない場合、usecase を呼ばず 400 を返す(Update/Delete)。
	called := false
	m := &mockTimeEntryUsecase{
		updateFn: func(_ context.Context, _ int64, _ usecase.TimeEntryInput) (*domain.TimeEntry, error) {
			called = true
			return nil, nil
		},
		deleteFn: func(_ context.Context, _ int64) error {
			called = true
			return nil
		},
	}
	h := newTimeEntryTestServer(m)

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		rec := do(t, h, method, "/time-entries/abc", `{"task_id":1,"user_id":2,"minutes":1,"spent_on":"2026-07-02"}`)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: want 400, got %d", method, rec.Code)
		}
	}
	if called {
		t.Error("usecase should not be called on invalid id")
	}
}

func TestTimeEntryHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func(m *mockTimeEntryUsecase)
		wantStatus int
	}{
		{
			name:       "正常系: 204",
			path:       "/time-entries/1",
			setupMock:  func(m *mockTimeEntryUsecase) { m.deleteFn = func(_ context.Context, _ int64) error { return nil } },
			wantStatus: http.StatusNoContent,
		},
		{
			name: "異常系: ErrNotFound を 404 に変換",
			path: "/time-entries/999",
			setupMock: func(m *mockTimeEntryUsecase) {
				m.deleteFn = func(_ context.Context, _ int64) error { return domain.ErrNotFound }
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockTimeEntryUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m)
			}
			h := newTimeEntryTestServer(m)

			rec := do(t, h, http.MethodDelete, tt.path, "")

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
