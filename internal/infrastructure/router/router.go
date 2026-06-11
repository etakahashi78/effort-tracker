// Package router はHTTPルーティングとミドルウェアを組み立てる(フレームワーク&ドライバ層)。
package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/etakahashi78/effort-tracker/internal/adapter/handler"
)

// New は各ハンドラを受け取り、全エンドポイントを集約した http.Handler を構築する。
func New(logger *slog.Logger, project *handler.ProjectHandler) http.Handler {
	// ---- ルーティング定義(全エンドポイントをここに集約) ----
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthz)

	// Project
	mux.HandleFunc("POST /projects", project.Create)
	mux.HandleFunc("GET /projects", project.List)
	mux.HandleFunc("GET /projects/{id}", project.Get)
	mux.HandleFunc("PUT /projects/{id}", project.Update)
	mux.HandleFunc("DELETE /projects/{id}", project.Delete)

	// TODO: Task / TimeEntry / User を同様にここへ追加する。

	return logging(logger, mux)
}

// healthz はヘルスチェックを返す。
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// logging は各リクエストのメソッド・パス・ステータス・所要時間を記録する。
func logging(logger *slog.Logger, next http.Handler) http.Handler {
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

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
