package database

import (
	"database/sql"
	"fmt"

	"github.com/camphor-/relaym-server/config"

	"gopkg.in/gorp.v1"
)

// NewDB はMySQLへ接続し、gorpのマッピングオブジェクトを生成します。
func NewDB() (*gorp.DbMap, error) {
	db, err := sql.Open("mysql", config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL: %w", err)
	}

	db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(100)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}

	return dbMap, nil
}
