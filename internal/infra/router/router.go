// Package router はHTTPルーティングとミドルウェアを組み立てる(フレームワーク&ドライバ層)。
package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/etakahashi78/effort-tracker/internal/adapter/handler"
)

// New は各ハンドラを受け取り、全エンドポイントを集約した http.Handler を構築する。
func New(logger *slog.Logger, project *handler.ProjectHandler, timeEntry *handler.TimeEntryHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(requestLogger(logger))

	// ---- ルーティング定義(全エンドポイントをここに集約) ----
	r.Get("/healthz", healthz)

	// Project
	r.Route("/projects", func(r chi.Router) {
		r.Post("/", project.Create)
		r.Get("/", project.List)
		r.Get("/{id}", project.Get)
		r.Put("/{id}", project.Update)
		r.Delete("/{id}", project.Delete)
	})

	// TimeEntry
	r.Route("/time-entries", func(r chi.Router) {
		r.Post("/", timeEntry.Create)
		r.Get("/", timeEntry.List)
		r.Get("/{id}", timeEntry.Get)
		r.Put("/{id}", timeEntry.Update)
		r.Delete("/{id}", timeEntry.Delete)
	})

	// TODO: Task / User を同様にここへ追加する。

	return r
}

// healthz はヘルスチェックを返す。
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// requestLogger は各リクエストのメソッド・パス・ステータス・所要時間を記録する chi ミドルウェア。
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration", time.Since(start).String(),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
