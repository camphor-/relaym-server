package database

import (
	"testing"

	"github.com/go-gorp/gorp/v3"
)

func truncateTable(t *testing.T, dbMap *gorp.DbMap) {
	t.Helper()

	if _, err := dbMap.Exec("set foreign_key_checks = 0"); err != nil {
		t.Fatal(err)
	}
	if err := dbMap.TruncateTables(); err != nil {
		t.Fatal(err)
	}
	if _, err := dbMap.Exec("set foreign_key_checks = 1"); err != nil {
		t.Fatal(err)
	}
}
