package database

import (
	"context"
	"database/sql"

	"github.com/go-gorp/gorp/v3"
)

var txKey = struct{}{}

type TransactionDAO interface {
	SelectOne(holder interface{}, query string, args ...interface{}) error
	Insert(list ...interface{}) error
	Update(list ...interface{}) (int64, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func getTx(ctx context.Context) (TransactionDAO, bool) {
	tx, ok := ctx.Value(&txKey).(*gorp.Transaction)
	return tx, ok
}
