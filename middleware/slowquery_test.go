package middleware

import (
	"context"
	"github.com/KNICEX/go-orm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewSlowQueryBuilder(t *testing.T) {
	builder := NewSlowQueryBuilder(time.Millisecond * 5)

	db, err := orm.Open("sqlite3", "file:test.db?cache=shared&mode=memory",
		orm.DBWithDialect(orm.DialectSQLite3), orm.DBWithMiddlewares(builder.Build()))
	require.NoError(t, err)

	_, _ = orm.NewSelector[TestModel](db).Where(orm.Col("Id").Eq(12)).Get(context.Background())

}
