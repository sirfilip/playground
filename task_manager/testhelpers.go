package task_manager

import (
	"database/sql"
	"embed"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func assertEquals(t *testing.T, a any, b any) {
	t.Helper()
	if a != b {
		t.Errorf("expected %v to be equal to %v", a, b)
	}
}

func assert(t *testing.T, cond bool, msg string) {
	t.Helper()
	if !cond {
		t.Error(msg)
	}
}

func connect(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatal(err)
	}
	return db
}
