package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/etakahashi78/effort-tracker/internal/adapter/handler"
	"github.com/etakahashi78/effort-tracker/internal/domain"
	"github.com/etakahashi78/effort-tracker/internal/infra/router"
	"github.com/etakahashi78/effort-tracker/internal/usecase"
)

// mockProjectUsecase は handler.ProjectUsecase の手書きモック。
type mockProjectUsecase struct {
	createFn func(ctx context.Context, in usecase.ProjectInput) (*domain.Project, error)
	listFn   func(ctx context.Context) ([]domain.Project, error)
	getFn    func(ctx context.Context, id int64) (*domain.Project, error)
	updateFn func(ctx context.Context, id int64, in usecase.ProjectInput) (*domain.Project, error)
	deleteFn func(ctx context.Context, id int64) error
}

func (m *mockProjectUsecase) Create(ctx context.Context, in usecase.ProjectInput) (*domain.Project, error) {
	return m.createFn(ctx, in)
}
func (m *mockProjectUsecase) List(ctx context.Context) ([]domain.Project, error) {
	return m.listFn(ctx)
}
func (m *mockProjectUsecase) Get(ctx context.Context, id int64) (*domain.Project, error) {
	return m.getFn(ctx, id)
}
func (m *mockProjectUsecase) Update(ctx context.Context, id int64, in usecase.ProjectInput) (*domain.Project, error) {
	return m.updateFn(ctx, id, in)
}
func (m *mockProjectUsecase) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

// newTestServer はモックを注入したハンドラを実ルータに載せた http.Handler を返す。
// chi の URL パラメータ解決まで含めて検証できる。
func newTestServer(uc handler.ProjectUsecase) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return router.New(logger, handler.NewProjectHandler(uc), handler.NewTimeEntryHandler(nil))
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestProjectHandler_Create(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		setupMock     func(m *mockProjectUsecase, gotIn *usecase.ProjectInput, called *bool)
		wantStatus    int
		checkInput    func(t *testing.T, gotIn usecase.ProjectInput, called bool)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 201 と作成結果のJSONを返す",
			body: `{"name":"PJ","description":"d"}`,
			setupMock: func(m *mockProjectUsecase, gotIn *usecase.ProjectInput, _ *bool) {
				m.createFn = func(_ context.Context, in usecase.ProjectInput) (*domain.Project, error) {
					*gotIn = in
					return &domain.Project{ID: 1, Name: in.Name, Status: "active"}, nil
				}
			},
			wantStatus: http.StatusCreated,
			checkInput: func(t *testing.T, gotIn usecase.ProjectInput, _ bool) {
				if gotIn.Name != "PJ" || gotIn.Description != "d" {
					t.Errorf("usecase received unexpected input: %+v", gotIn)
				}
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got domain.Project
				mustDecode(t, rec, &got)
				if got.ID != 1 {
					t.Errorf("want id 1, got %d", got.ID)
				}
			},
		},
		{
			name: "異常系: 不正なJSONは usecase を呼ばず 400",
			body: `{`,
			setupMock: func(m *mockProjectUsecase, _ *usecase.ProjectInput, called *bool) {
				m.createFn = func(_ context.Context, _ usecase.ProjectInput) (*domain.Project, error) {
					*called = true
					return nil, nil
				}
			},
			wantStatus: http.StatusBadRequest,
			checkInput: func(t *testing.T, _ usecase.ProjectInput, called bool) {
				if called {
					t.Error("usecase should not be called on bad JSON")
				}
			},
		},
		{
			name: "異常系: usecase の ErrInvalidInput を 400 に変換",
			body: `{}`,
			setupMock: func(m *mockProjectUsecase, _ *usecase.ProjectInput, _ *bool) {
				m.createFn = func(_ context.Context, _ usecase.ProjectInput) (*domain.Project, error) {
					return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotIn usecase.ProjectInput
			called := false
			m := &mockProjectUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m, &gotIn, &called)
			}
			h := newTestServer(m)

			rec := do(t, h, http.MethodPost, "/projects", tt.body)

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

func TestProjectHandler_Get(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		setupMock     func(m *mockProjectUsecase, called *bool)
		wantStatus    int
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 200",
			path: "/projects/5",
			setupMock: func(m *mockProjectUsecase, _ *bool) {
				m.getFn = func(_ context.Context, id int64) (*domain.Project, error) {
					return &domain.Project{ID: id, Name: "PJ"}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got domain.Project
				mustDecode(t, rec, &got)
				if got.ID != 5 {
					t.Errorf("chi URL param not resolved: want id 5, got %d", got.ID)
				}
			},
		},
		{
			name: "異常系: ErrNotFound を 404 に変換",
			path: "/projects/999",
			setupMock: func(m *mockProjectUsecase, _ *bool) {
				m.getFn = func(_ context.Context, _ int64) (*domain.Project, error) {
					return nil, domain.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "異常系: 数値でない id は usecase を呼ばず 400",
			path: "/projects/abc",
			setupMock: func(m *mockProjectUsecase, called *bool) {
				m.getFn = func(_ context.Context, _ int64) (*domain.Project, error) {
					*called = true
					return nil, nil
				}
			},
			wantStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// called check is done in the main loop if we pass the flag
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			m := &mockProjectUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m, &called)
			}
			h := newTestServer(m)

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

func TestProjectHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func(m *mockProjectUsecase, gotID *int64, gotIn *usecase.ProjectInput, called *bool)
		wantStatus int
		checkInput func(t *testing.T, gotID int64, gotIn usecase.ProjectInput, called bool)
	}{
		{
			name: "正常系: 200、id とボディを usecase に渡す",
			path: "/projects/8",
			body: `{"name":"改","status":"archived"}`,
			setupMock: func(m *mockProjectUsecase, gotID *int64, gotIn *usecase.ProjectInput, _ *bool) {
				m.updateFn = func(_ context.Context, id int64, in usecase.ProjectInput) (*domain.Project, error) {
					*gotID, *gotIn = id, in
					return &domain.Project{ID: id, Name: in.Name, Status: in.Status}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkInput: func(t *testing.T, gotID int64, gotIn usecase.ProjectInput, _ bool) {
				if gotID != 8 {
					t.Errorf("want id 8, got %d", gotID)
				}
				if gotIn.Name != "改" || gotIn.Status != "archived" {
					t.Errorf("usecase received unexpected input: %+v", gotIn)
				}
			},
		},
		{
			name: "異常系: ErrNotFound を 404 に変換",
			path: "/projects/999",
			body: `{"name":"x"}`,
			setupMock: func(m *mockProjectUsecase, _ *int64, _ *usecase.ProjectInput, _ *bool) {
				m.updateFn = func(_ context.Context, _ int64, _ usecase.ProjectInput) (*domain.Project, error) {
					return nil, domain.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "異常系: 不正なJSONは 400",
			path: "/projects/1",
			body: `{bad`,
			setupMock: func(m *mockProjectUsecase, _ *int64, _ *usecase.ProjectInput, called *bool) {
				m.updateFn = func(_ context.Context, _ int64, _ usecase.ProjectInput) (*domain.Project, error) {
					*called = true
					return nil, nil
				}
			},
			wantStatus: http.StatusBadRequest,
			checkInput: func(t *testing.T, _ int64, _ usecase.ProjectInput, called bool) {
				if called {
					t.Error("usecase should not be called on bad JSON")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID int64
			var gotIn usecase.ProjectInput
			called := false
			m := &mockProjectUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m, &gotID, &gotIn, &called)
			}
			h := newTestServer(m)

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

func TestProjectHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mockProjectUsecase)
		wantStatus int
		checkBody  func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 200",
			setupMock: func(m *mockProjectUsecase) {
				m.listFn = func(_ context.Context) ([]domain.Project, error) {
					return []domain.Project{{ID: 1}, {ID: 2}}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got []domain.Project
				mustDecode(t, rec, &got)
				if len(got) != 2 {
					t.Errorf("want 2 projects, got %d", len(got))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockProjectUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m)
			}
			h := newTestServer(m)

			rec := do(t, h, http.MethodGet, "/projects", "")

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d", tt.wantStatus, rec.Code)
			}
			if tt.checkBody != nil {
				tt.checkBody(t, rec)
			}
		})
	}
}

func TestProjectHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func(m *mockProjectUsecase)
		wantStatus int
	}{
		{
			name: "正常系: 204",
			path: "/projects/1",
			setupMock: func(m *mockProjectUsecase) {
				m.deleteFn = func(_ context.Context, _ int64) error { return nil }
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "異常系: ErrNotFound を 404 に変換",
			path: "/projects/999",
			setupMock: func(m *mockProjectUsecase) {
				m.deleteFn = func(_ context.Context, _ int64) error { return domain.ErrNotFound }
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockProjectUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m)
			}
			h := newTestServer(m)

			rec := do(t, h, http.MethodDelete, tt.path, "")

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func mustDecode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
