package middleware

import (
	"context"
	"github.com/KNICEX/go-orm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestModel struct {
	Id        int
	FirstName string
	LastName  string
}

func TestNewLogBuilder(t *testing.T) {
	var query string
	var args []any
	b := NewLogBuilder().LogFunc(func(q string, a []any) {
		query = q
		args = a
	})
	db, err := orm.Open("sqlite3", "file:test.db?cache=shared&mode=memory",
		orm.DBWithDialect(orm.DialectSQLite3), orm.DBWithMiddlewares(b.Build()))
	require.NoError(t, err)

	_, _ = orm.NewSelector[TestModel](db).Where(orm.Col("Id").Eq(12)).Get(context.Background())
	assert.Equal(t, "SELECT * FROM `test_model` WHERE `id` = ? LIMIT 1;", query)
	assert.Equal(t, []any{12}, args)

	_ = orm.NewInserter[TestModel](db).Values(&TestModel{Id: 12, FirstName: "first", LastName: "last"}).Exec(context.Background())
	assert.Equal(t, "INSERT INTO `test_model` (`id`,`first_name`,`last_name`) VALUES (?,?,?);", query)
	assert.Equal(t, []any{12, "first", "last"}, args)

	_ = orm.NewUpdater[TestModel](db).Set(orm.Assign("FirstName", "Tom")).Where(orm.Col("Id").Eq(12)).Exec(context.Background())
	assert.Equal(t, "UPDATE `test_model` SET `first_name` = ? WHERE `id` = ?;", query)
	assert.Equal(t, []any{"Tom", 12}, args)

	_ = orm.NewDeleter[TestModel](db).Where(orm.Col("Id").Eq(12)).Exec(context.Background())
	assert.Equal(t, "DELETE FROM `test_model` WHERE `id` = ?;", query)
	assert.Equal(t, []any{12}, args)
}
