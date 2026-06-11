// Package handler はHTTP ⇔ usecase の変換を担う(インターフェースアダプタ層)。
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/etakahashi78/effort-tracker/internal/domain"
)

// writeJSON は値をJSONとして指定ステータスで書き出す。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// writeError はエラーメッセージをJSONで返す。
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// mapError はドメイン/ユースケース層のエラーを適切なHTTPレスポンスに変換する。
// レスポンスを書き出した場合は true を返す。
func mapError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
		return true
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
		return true
	case errors.Is(err, context.Canceled):
		writeError(w, http.StatusRequestTimeout, "request canceled")
		return true
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
		return true
	}
}
