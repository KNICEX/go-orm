package orm

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

func memoryDB(t *testing.T, opts ...DBOption) *DB {
	db, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory", opts...)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func memoryWithDB(db string, t *testing.T, opts ...DBOption) *DB {
	orm, err := Open("sqlite3", fmt.Sprintf("file:%s.db?cache=shared&mode=memory", db), opts...)
	if err != nil {
		t.Fatal(err)
	}
	return orm

}
