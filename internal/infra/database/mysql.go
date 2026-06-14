// Package database はMySQLへの接続とスキーマ初期化を提供する(フレームワーク&ドライバ層)。
package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

//go:embed schema.sql
var schema string

// Open は指定DSNでMySQLに接続し、スキーマを適用して *sql.DB を返す。
// dsn は go-sql-driver/mysql 形式: "user:pass@tcp(host:port)/dbname"。
// ParseTime / MultiStatements はアプリ要件として強制的に有効化する。
func Open(dsn string) (*sql.DB, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	cfg.ParseTime = true       // DATETIME/TIMESTAMP を time.Time にスキャン
	cfg.MultiStatements = true // schema.sql を1回のExecで適用するため
	cfg.Loc = time.UTC

	conn, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	conn.SetConnMaxLifetime(3 * time.Minute)
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(10)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	return conn, nil
}
