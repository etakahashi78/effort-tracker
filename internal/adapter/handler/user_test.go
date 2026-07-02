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

// mockUserUsecase は handler.UserUsecase の手書きモック。
type mockUserUsecase struct {
	createFn func(ctx context.Context, in usecase.UserInput) (*domain.User, error)
	listFn   func(ctx context.Context) ([]domain.User, error)
	getFn    func(ctx context.Context, id int64) (*domain.User, error)
	updateFn func(ctx context.Context, id int64, in usecase.UserInput) (*domain.User, error)
	deleteFn func(ctx context.Context, id int64) error
}

func (m *mockUserUsecase) Create(ctx context.Context, in usecase.UserInput) (*domain.User, error) {
	return m.createFn(ctx, in)
}
func (m *mockUserUsecase) List(ctx context.Context) ([]domain.User, error) {
	return m.listFn(ctx)
}
func (m *mockUserUsecase) Get(ctx context.Context, id int64) (*domain.User, error) {
	return m.getFn(ctx, id)
}
func (m *mockUserUsecase) Update(ctx context.Context, id int64, in usecase.UserInput) (*domain.User, error) {
	return m.updateFn(ctx, id, in)
}
func (m *mockUserUsecase) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

// newUserTestServer はモックを注入したハンドラを実ルータに載せた http.Handler を返す。
func newUserTestServer(uc handler.UserUsecase) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return router.New(logger, handler.NewProjectHandler(nil), handler.NewTimeEntryHandler(nil), handler.NewUserHandler(uc))
}

func userDo(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
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

func TestUserHandler_Create(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		setupMock     func(m *mockUserUsecase, gotIn *usecase.UserInput, called *bool)
		wantStatus    int
		checkInput    func(t *testing.T, gotIn usecase.UserInput, called bool)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 201 と作成結果のJSONを返す",
			body: `{"name":"太郎","email":"taro@example.com"}`,
			setupMock: func(m *mockUserUsecase, gotIn *usecase.UserInput, _ *bool) {
				m.createFn = func(_ context.Context, in usecase.UserInput) (*domain.User, error) {
					*gotIn = in
					return &domain.User{ID: 1, Name: in.Name, Email: in.Email}, nil
				}
			},
			wantStatus: http.StatusCreated,
			checkInput: func(t *testing.T, gotIn usecase.UserInput, _ bool) {
				if gotIn.Name != "太郎" || gotIn.Email != "taro@example.com" {
					t.Errorf("usecase received unexpected input: %+v", gotIn)
				}
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got domain.User
				userMustDecode(t, rec, &got)
				if got.ID != 1 {
					t.Errorf("want id 1, got %d", got.ID)
				}
			},
		},
		{
			name: "異常系: 不正なJSONは usecase を呼ばず 400",
			body: `{`,
			setupMock: func(m *mockUserUsecase, _ *usecase.UserInput, called *bool) {
				m.createFn = func(_ context.Context, _ usecase.UserInput) (*domain.User, error) {
					*called = true
					return nil, nil
				}
			},
			wantStatus: http.StatusBadRequest,
			checkInput: func(t *testing.T, _ usecase.UserInput, called bool) {
				if called {
					t.Error("usecase should not be called on bad JSON")
				}
			},
		},
		{
			name: "異常系: usecase の ErrInvalidInput を 400 に変換",
			body: `{}`,
			setupMock: func(m *mockUserUsecase, _ *usecase.UserInput, _ *bool) {
				m.createFn = func(_ context.Context, _ usecase.UserInput) (*domain.User, error) {
					return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
				}
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotIn usecase.UserInput
			called := false
			m := &mockUserUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m, &gotIn, &called)
			}
			h := newUserTestServer(m)

			rec := userDo(t, h, http.MethodPost, "/users", tt.body)

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

func TestUserHandler_Get(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		setupMock     func(m *mockUserUsecase, called *bool)
		wantStatus    int
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 200",
			path: "/users/5",
			setupMock: func(m *mockUserUsecase, _ *bool) {
				m.getFn = func(_ context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: id, Name: "太郎"}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got domain.User
				userMustDecode(t, rec, &got)
				if got.ID != 5 {
					t.Errorf("chi URL param not resolved: want id 5, got %d", got.ID)
				}
			},
		},
		{
			name: "異常系: ErrNotFound を 404 に変換",
			path: "/users/999",
			setupMock: func(m *mockUserUsecase, _ *bool) {
				m.getFn = func(_ context.Context, _ int64) (*domain.User, error) {
					return nil, domain.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "異常系: 数値でない id は usecase を呼ばず 400",
			path: "/users/abc",
			setupMock: func(m *mockUserUsecase, called *bool) {
				m.getFn = func(_ context.Context, _ int64) (*domain.User, error) {
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
			m := &mockUserUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m, &called)
			}
			h := newUserTestServer(m)

			rec := userDo(t, h, http.MethodGet, tt.path, "")

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d (%s)", tt.wantStatus, rec.Code, rec.Body)
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

func TestUserHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		body       string
		setupMock  func(m *mockUserUsecase, gotID *int64, gotIn *usecase.UserInput, called *bool)
		wantStatus int
		checkInput func(t *testing.T, gotID int64, gotIn usecase.UserInput, called bool)
	}{
		{
			name: "正常系: 200",
			path: "/users/3",
			body: `{"name":"太郎","email":"new@example.com"}`,
			setupMock: func(m *mockUserUsecase, gotID *int64, gotIn *usecase.UserInput, _ *bool) {
				m.updateFn = func(_ context.Context, id int64, in usecase.UserInput) (*domain.User, error) {
					*gotID = id
					*gotIn = in
					return &domain.User{ID: id, Name: in.Name, Email: in.Email}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkInput: func(t *testing.T, gotID int64, gotIn usecase.UserInput, _ bool) {
				if gotID != 3 {
					t.Errorf("want id 3, got %d", gotID)
				}
				if gotIn.Name != "太郎" || gotIn.Email != "new@example.com" {
					t.Errorf("unexpected input: %+v", gotIn)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID int64
			var gotIn usecase.UserInput
			called := false
			m := &mockUserUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m, &gotID, &gotIn, &called)
			}
			h := newUserTestServer(m)

			rec := userDo(t, h, http.MethodPut, tt.path, tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d (%s)", tt.wantStatus, rec.Code, rec.Body)
			}
			if tt.checkInput != nil {
				tt.checkInput(t, gotID, gotIn, called)
			}
		})
	}
}

func TestUserHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(m *mockUserUsecase)
		wantStatus int
		checkBody  func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "正常系: 200",
			setupMock: func(m *mockUserUsecase) {
				m.listFn = func(_ context.Context) ([]domain.User, error) {
					return []domain.User{{ID: 1}, {ID: 2}}, nil
				}
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var got []domain.User
				userMustDecode(t, rec, &got)
				if len(got) != 2 {
					t.Errorf("want 2 users, got %d", len(got))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockUserUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m)
			}
			h := newUserTestServer(m)

			rec := userDo(t, h, http.MethodGet, "/users", "")

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d", tt.wantStatus, rec.Code)
			}
			if tt.checkBody != nil {
				tt.checkBody(t, rec)
			}
		})
	}
}

func TestUserHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func(m *mockUserUsecase)
		wantStatus int
	}{
		{
			name: "正常系: 204",
			path: "/users/1",
			setupMock: func(m *mockUserUsecase) {
				m.deleteFn = func(_ context.Context, _ int64) error { return nil }
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "異常系: ErrNotFound を 404 に変換",
			path: "/users/999",
			setupMock: func(m *mockUserUsecase) {
				m.deleteFn = func(_ context.Context, _ int64) error { return domain.ErrNotFound }
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockUserUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(m)
			}
			h := newUserTestServer(m)

			rec := userDo(t, h, http.MethodDelete, tt.path, "")

			if rec.Code != tt.wantStatus {
				t.Fatalf("want %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func userMustDecode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
