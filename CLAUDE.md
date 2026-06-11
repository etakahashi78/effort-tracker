# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

工数管理ツール (effort-tracker) の REST API プロトタイプ。Go 1.26 / 標準ライブラリ `net/http` / MySQL 8 (`github.com/go-sql-driver/mysql`)。

## Commands

```bash
docker compose up -d     # ローカル開発用 MySQL (+ Adminer) 起動。先に必須
go run ./cmd/server      # 起動 (ADDR=:8080, DB_DSN=app:app@tcp(127.0.0.1:3306)/effort_tracker)
go build ./...           # ビルド
go vet ./...             # 静的解析
go test ./...            # テスト (現状テストは未整備)
docker compose down [-v] # MySQL停止 (-v でデータ破棄)
```

環境変数: `ADDR`（リッスンアドレス, 既定 `:8080`）、`DB_DSN`（MySQL DSN = go-sql-driver形式, 既定 `app:app@tcp(127.0.0.1:3306)/effort_tracker`）。

`compose.yaml` が MySQL 8.4（`localhost:3306`, db=`effort_tracker`/user=`app`/pass=`app`）と Adminer（http://localhost:8081）を提供する。アプリ本体はホスト側で `go run` し、コンテナのMySQLへ接続する。

## Architecture

クリーンアーキテクチャ。**依存方向は常に内向き** (`infrastructure` / `adapter` → `usecase` → `domain`)。最内層の `domain` は何にも依存しない。永続化は `domain` が定義するインターフェース(ポート)を外側の `adapter/persistence` が実装する**依存性逆転**で結合する。

```
internal/
├── domain/          # 最内層: エンティティ + リポジトリIF + ドメインエラー (依存なし)
├── usecase/         # アプリケーションビジネスルール (バリデーション・既定値)
├── adapter/
│   ├── handler/     # HTTP ⇔ usecase の変換 (インターフェースアダプタ)
│   └── persistence/ # domain リポジトリIF の SQLite 実装
└── infrastructure/
    ├── database/    # MySQL接続 + schema.sql (フレームワーク&ドライバ)
    └── router/      # ルーティング集約 + logging ミドルウェア
```

- **`cmd/server/main.go`** — 合成ルート(composition root)。具象実装をインターフェースへ配線する唯一の場所: `persistence` → `usecase` → `handler` → `router` の順に注入。env読み込み・グレースフルシャットダウンもここ。
- **`internal/domain`** — エンティティ `User`/`Project`/`Task`/`TimeEntry`(`model.go`)、リポジトリのポート(`repository.go`)、共通エラー `ErrNotFound`/`ErrInvalidInput`(`errors.go`)。
- **`internal/usecase`** — `ProjectUsecase` が `domain.ProjectRepository` インターフェースにのみ依存。バリデーションと既定値設定(`buildProject`)はここ。違反は `fmt.Errorf("%w: ...", domain.ErrInvalidInput)` でラップして返す。DB/HTTPを知らない。
- **`internal/adapter/persistence`** — `domain.ProjectRepository` のMySQL実装。`var _ domain.ProjectRepository = (*ProjectRepository)(nil)` で実装漏れをコンパイル時検出。レコード無しは `domain.ErrNotFound`。**MySQLは `RETURNING` 非対応**のため、Create は `LastInsertId` 取得後に `Get` で再取得、Update は `Exec` 後に `Get` で再取得(存在しなければ `Get` が `ErrNotFound`)。
- **`internal/adapter/handler`** — HTTPハンドラ。消費側で `ProjectUsecase` インターフェースを定義(テスト容易性)。`writeJSON`/`writeError`、`mapError` がエラーをHTTPへ変換(`ErrInvalidInput`→400, `ErrNotFound`→404, `context.Canceled`→408, その他→500)。
- **`internal/infrastructure/database`** — `Open(dsn)` がMySQLに接続し `schema.sql`(`//go:embed`)を適用。`mysql.ParseDSN` で `ParseTime`/`MultiStatements`/`Loc=UTC` を強制有効化(複数文スキーマを1 Execで適用するため)。`CREATE TABLE IF NOT EXISTS` で起動毎に冪等適用。FKがInnoDB索引を自動生成するため明示的 `CREATE INDEX` は持たない(MySQLは `CREATE INDEX IF NOT EXISTS` 非対応のため)。マイグレーションツールは無し。
- **`internal/infrastructure/router`** — `New(logger, *handler.ProjectHandler, ...)` が全エンドポイントのルート定義を**一箇所に集約**して `http.ServeMux` に登録し `logging` で包む。

ルーティングは Go 1.22+ の `ServeMux` パターン（例: `"GET /projects/{id}"`、パス変数は `r.PathValue("id")`）。ログは `slog` のJSONハンドラ。

## 実装状況と拡張パターン

現在 **Project** のみ CRUD 実装済み。`User` / `Task` / `TimeEntry` はエンティティとスキーマは定義済みだが usecase/persistence/handler 未実装。

新エンティティを追加する際は Project と同じ垂直スライスに従う(内側から外側へ):
1. `internal/domain/repository.go` に `<Entity>Repository` インターフェースを追加
2. `internal/usecase/<entity>.go` に `<Entity>Usecase`(`New...`, バリデーション含む CRUD)と `<Entity>Input`
3. `internal/adapter/persistence/<entity>.go` に SQLite 実装(`var _ domain.<Entity>Repository = ...` で検証)
4. `internal/adapter/handler/<entity>.go` に `<Entity>Handler`・消費側インターフェース・`<entity>Input`・`toUsecase()`
5. `internal/infrastructure/router/router.go` の `New` 引数とルート定義ブロックにハンドラを追加
6. `cmd/server/main.go` の合成ルートで配線

ハンドラの慣習: ボディは `json.Decoder` + `DisallowUnknownFields()` でデコード、`toUsecase()` で usecase 入力へ変換、パスIDは `pathID(r)` ヘルパで取得。バリデーションは usecase 層で行い handler は持たない。
