package sql_demo

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

var createTableSql = `
CREATE TABLE IF NOT EXISTS test (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	first_name TEXT NOT NULL,
    	last_name TEXT NOT NULL
);
`

func TestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)

	require.NoError(t, db.Ping())
	defer func() {
		_ = db.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, createTableSql)
	require.NoError(t, err)

	res, err := db.ExecContext(ctx, "INSERT INTO test (first_name, last_name) VALUES (?, ?)", "foo", "bar")
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), affected)
	log.Println("affected:", affected)
	lastId, err := res.LastInsertId()
	require.NoError(t, err)
	log.Println("lastId:", lastId)

	row := db.QueryRowContext(ctx, "SELECT id, first_name, last_name FROM test WHERE id = ?", lastId)
	require.NoError(t, row.Err())
	tm := TestModel{}
	err = row.Scan(&tm.Id, &tm.FirstName, &tm.LastName)
	require.NoError(t, err)
	log.Println(tm)

	// 无查询结果
	row = db.QueryRowContext(ctx, "SELECT id, first_name, last_name FROM test WHERE id = ?", 2)
	require.NoError(t, row.Err())

	err = row.Scan(&tm.Id, &tm.FirstName, &tm.LastName)
	require.Error(t, err)

}

type TestModel struct {
	Id        int64
	LastName  string
	FirstName string
}
