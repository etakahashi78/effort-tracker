package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/etakahashi78/effort-tracker/internal/domain"
	"github.com/etakahashi78/effort-tracker/internal/usecase"
)

// ProjectUsecase は handler が必要とするユースケースの契約(消費側で定義)。
type ProjectUsecase interface {
	Create(ctx context.Context, in usecase.ProjectInput) (*domain.Project, error)
	List(ctx context.Context) ([]domain.Project, error)
	Get(ctx context.Context, id int64) (*domain.Project, error)
	Update(ctx context.Context, id int64, in usecase.ProjectInput) (*domain.Project, error)
	Delete(ctx context.Context, id int64) error
}

// ProjectHandler は /projects 系エンドポイントを処理する。
type ProjectHandler struct {
	uc ProjectUsecase
}

// NewProjectHandler は ProjectHandler を生成する。
func NewProjectHandler(uc ProjectUsecase) *ProjectHandler {
	return &ProjectHandler{uc: uc}
}

// projectInput はリクエストボディの受け口。
type projectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (in projectInput) toUsecase() usecase.ProjectInput {
	return usecase.ProjectInput{Name: in.Name, Description: in.Description, Status: in.Status}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	in, err := decodeProjectInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.uc.Create(r.Context(), in.toUsecase())
	if mapError(w, err) {
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.uc.List(r.Context())
	if mapError(w, err) {
		return
	}
	if projects == nil {
		projects = []domain.Project{}
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	p, err := h.uc.Get(r.Context(), id)
	if mapError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	in, err := decodeProjectInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.uc.Update(r.Context(), id, in.toUsecase())
	if mapError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func decodeProjectInput(r *http.Request) (projectInput, error) {
	var in projectInput
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return in, errors.New("invalid JSON body: " + err.Error())
	}
	return in, nil
}

// pathID は URL パス変数 {id} を int64 として取り出す。
func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
