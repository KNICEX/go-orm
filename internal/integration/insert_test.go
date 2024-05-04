//go:build e2e

package integration

import (
	"context"
	"github.com/KNICEX/go-orm"
	"github.com/KNICEX/go-orm/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

type InsertSuite struct {
	Suite
}

func (i *InsertSuite) SetupSuite() {
	i.Suite.SetupSuite()
	err := orm.RawQuery[any](i.db, "TRUNCATE TABLE simple_struct;").Exec(context.Background()).Err()
	assert.NoError(i.T(), err)

}

func TestMySQLInsert(t *testing.T) {
	suite.Run(t, &InsertSuite{
		Suite{
			driver:  "mysql",
			dialect: orm.DialectMySQL,
			dsn:     "root:root@tcp(localhost:3307)/integration_test",
		},
	})
}

func (i *InsertSuite) TearDownTest() {
	err := orm.RawQuery[any](i.db, "TRUNCATE TABLE simple_struct;").Exec(context.Background()).Err()
	assert.NoError(i.T(), err)
}

func (i *InsertSuite) TestInsert() {
	db := i.db
	t := i.T()
	testCases := []struct {
		name         string
		i            *orm.Inserter[test.SimpleStruct]
		wantAffected int64
	}{
		{
			name:         "insert one",
			i:            orm.NewInserter[test.SimpleStruct](db).Values(test.NewSimpleStruct(1)),
			wantAffected: 1,
		},
		{
			name: "insert multiple",
			i: orm.NewInserter[test.SimpleStruct](db).Values(
				test.NewSimpleStruct(2),
				test.NewSimpleStruct(3),
			),
			wantAffected: 2,
		},
		{
			name:         "just id",
			i:            orm.NewInserter[test.SimpleStruct](db).Values(&test.SimpleStruct{Id: 4}),
			wantAffected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.i.Exec(context.Background())
			assert.NoError(t, res.Err())
			affected, err := res.RowsAffected()
			assert.NoError(t, err)
			assert.Equal(t, tc.wantAffected, affected)
		})
	}
}
