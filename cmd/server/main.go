// Command server は effort-tracker のREST APIサーバを起動する。
// ここが合成ルート(composition root): 具象実装をインターフェースへ配線する唯一の場所。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/etakahashi78/effort-tracker/internal/adapter/handler"
	"github.com/etakahashi78/effort-tracker/internal/adapter/persistence"
	"github.com/etakahashi78/effort-tracker/internal/infrastructure/database"
	"github.com/etakahashi78/effort-tracker/internal/infrastructure/router"
	"github.com/etakahashi78/effort-tracker/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr := envOr("ADDR", ":8080")
	dsn := envOr("DB_DSN", "app:app@tcp(127.0.0.1:3306)/effort_tracker")

	db, err := database.Open(dsn)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// ---- 依存の配線(外側→内側へ注入) ----
	projectRepo := persistence.NewProjectRepository(db) // domain.ProjectRepository を満たす
	projectUC := usecase.NewProjectUsecase(projectRepo)
	projectHandler := handler.NewProjectHandler(projectUC)

	srv := &http.Server{
		Addr:              addr,
		Handler:           router.New(logger, projectHandler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// グレースフルシャットダウン。
	go func() {
		logger.Info("server starting", "addr", addr, "dsn", dsn)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
