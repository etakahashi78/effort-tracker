# effort-tracker

工数管理ツールのREST API（プロトタイプ）。

## 技術スタック

- Go 1.26
- DB: MySQL 8（`github.com/go-sql-driver/mysql`）。ローカル開発は `compose.yaml` で起動

## ディレクトリ構成（クリーンアーキテクチャ）

```
effort-tracker/
├── cmd/server/main.go              # 合成ルート（配線・起動・グレースフルシャットダウン）
├── internal/
│   ├── domain/                     # 最内層: エンティティ + リポジトリIF + ドメインエラー
│   ├── usecase/                    # アプリケーションビジネスルール（バリデーション等）
│   ├── adapter/
│   │   ├── handler/                # HTTP ⇔ usecase の変換
│   │   └── persistence/            # domain リポジトリIF の MySQL 実装
│   └── infra/
│       ├── database/               # MySQL接続 + schema.sql（embed）
│       └── router/                 # ルーティング集約 + ミドルウェア
```

依存方向は常に内向き: `infra` / `adapter` → `usecase` → `domain`。
永続化は `domain` のインターフェースを `adapter/persistence` が実装する依存性逆転で結合。

## 実行

```bash
docker compose up -d          # MySQL を起動
go mod download
go run ./cmd/server           # ホスト側からMySQLへ接続
```

- MySQL: `localhost:3306`（db=`effort_tracker`, user=`app`, pass=`app`）
- 停止: `docker compose down`（データ保持） / `docker compose down -v`（データ破棄）

環境変数:

| 変数      | デフォルト                                          | 説明                          |
|-----------|-----------------------------------------------------|-------------------------------|
| `ADDR`    | `:8080`                                             | リッスンアドレス              |
| `DB_DSN`  | `app:app@tcp(127.0.0.1:3306)/effort_tracker`        | MySQL DSN（go-sql-driver形式）|

## API（Project / 実装済み）

| メソッド | パス              | 説明           |
|----------|-------------------|----------------|
| `GET`    | `/healthz`        | ヘルスチェック |
| `POST`   | `/projects`       | 作成           |
| `GET`    | `/projects`       | 一覧           |
| `GET`    | `/projects/{id}`  | 取得           |
| `PUT`    | `/projects/{id}`  | 更新           |
| `DELETE` | `/projects/{id}`  | 削除           |

### 例

```bash
# 作成
curl -X POST localhost:8080/projects \
  -H 'Content-Type: application/json' \
  -d '{"name":"社内ツール開発","description":"工数管理ツール"}'

# 一覧
curl localhost:8080/projects
```

## 次のステップ

`Project` と同じ垂直スライス（`domain` IF → `usecase` → `persistence` → `handler` →
`router`/`main` での配線）で `Task` → `TimeEntry` → `User` のCRUDを順に追加する。
詳細は `CLAUDE.md` の「実装状況と拡張パターン」を参照。
