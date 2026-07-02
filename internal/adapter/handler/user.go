package handler

import (
	"context"
	"net/http"

	"github.com/etakahashi78/effort-tracker/internal/domain"
	"github.com/etakahashi78/effort-tracker/internal/usecase"
)

// UserUsecase は handler が必要とするユースケースの契約(消費側で定義)。
type UserUsecase interface {
	Create(ctx context.Context, in usecase.UserInput) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Get(ctx context.Context, id int64) (*domain.User, error)
	Update(ctx context.Context, id int64, in usecase.UserInput) (*domain.User, error)
	Delete(ctx context.Context, id int64) error
}

// UserHandler は /users 系エンドポイントを処理する。
type UserHandler struct {
	uc UserUsecase
}

// NewUserHandler は UserHandler を生成する。
func NewUserHandler(uc UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

// userInput はリクエストボディの受け口。
type userInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (in userInput) toUsecase() usecase.UserInput {
	return usecase.UserInput{Name: in.Name, Email: in.Email}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in userInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.uc.Create(r.Context(), in.toUsecase())
	if mapError(w, err) {
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.uc.List(r.Context())
	if mapError(w, err) {
		return
	}
	if users == nil {
		users = []domain.User{}
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	u, err := h.uc.Get(r.Context(), id)
	if mapError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var in userInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.uc.Update(r.Context(), id, in.toUsecase())
	if mapError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.uc.Delete(r.Context(), id); mapError(w, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
