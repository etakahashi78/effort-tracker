package handler

import (
	"context"
	"net/http"

	"github.com/etakahashi78/effort-tracker/internal/domain"
	"github.com/etakahashi78/effort-tracker/internal/usecase"
)

// TimeEntryUsecase は handler が必要とするユースケースの契約(消費側で定義)。
type TimeEntryUsecase interface {
	Create(ctx context.Context, in usecase.TimeEntryInput) (*domain.TimeEntry, error)
	List(ctx context.Context) ([]domain.TimeEntry, error)
	Get(ctx context.Context, id int64) (*domain.TimeEntry, error)
	Update(ctx context.Context, id int64, in usecase.TimeEntryInput) (*domain.TimeEntry, error)
	Delete(ctx context.Context, id int64) error
}

// TimeEntryHandler は /time-entries 系エンドポイントを処理する。
type TimeEntryHandler struct {
	uc TimeEntryUsecase
}

// NewTimeEntryHandler は TimeEntryHandler を生成する。
func NewTimeEntryHandler(uc TimeEntryUsecase) *TimeEntryHandler {
	return &TimeEntryHandler{uc: uc}
}

// timeEntryInput はリクエストボディの受け口。
type timeEntryInput struct {
	TaskID  int64  `json:"task_id"`
	UserID  int64  `json:"user_id"`
	Minutes int    `json:"minutes"`
	Note    string `json:"note"`
	SpentOn string `json:"spent_on"`
}

func (in timeEntryInput) toUsecase() usecase.TimeEntryInput {
	return usecase.TimeEntryInput{
		TaskID:  in.TaskID,
		UserID:  in.UserID,
		Minutes: in.Minutes,
		Note:    in.Note,
		SpentOn: in.SpentOn,
	}
}

func (h *TimeEntryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in timeEntryInput
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

func (h *TimeEntryHandler) List(w http.ResponseWriter, r *http.Request) {
	entries, err := h.uc.List(r.Context())
	if mapError(w, err) {
		return
	}
	if entries == nil {
		entries = []domain.TimeEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *TimeEntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	e, err := h.uc.Get(r.Context(), id)
	if mapError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (h *TimeEntryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var in timeEntryInput
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

func (h *TimeEntryHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
