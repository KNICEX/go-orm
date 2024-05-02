package orm

import (
	"context"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInserter_SQLite_upsert_Build(t *testing.T) {
	db, err := OpenDB(nil, DBWithDialect(DialectSQLite3))
	require.NoError(t, err)
	testCases := []struct {
		name      string
		i         *Inserter[TestModel]
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "upsert with new value",
			i: NewInserter[TestModel](db).Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"}).
				OnDuplicateKey().ConflictColumns("Id").Update(Assign("FirstName", "newA"), Assign("LastName", "newB")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`,`last_name`) VALUES (?,?,?) " +
					"ON CONFLICT(`id`) DO UPDATE SET `first_name` = ?,`last_name` = ?;",
				Args: []any{int64(1), "a", "b", "newA", "newB"},
			},
		},
		{
			name: "upsert with column value",
			i: NewInserter[TestModel](db).Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"}).
				OnDuplicateKey().ConflictColumns("Id").Update(Col("FirstName"), Col("LastName")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`,`last_name`) VALUES (?,?,?) " +
					"ON CONFLICT(`id`) DO UPDATE SET `first_name` = EXCLUDED.`first_name`,`last_name` = EXCLUDED.`last_name`;",
				Args: []any{int64(1), "a", "b"},
			},
		},
		//{
		//	name: "upsert duplicated key without specified column",
		//	i:    NewInserter[TestModel](db).Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"}).OnDuplicateKey().Update(),
		//},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.i.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestInserter_MySQL_Build(t *testing.T) {
	db, err := OpenDB(nil, DBWithDialect(DialectMySQL))
	require.NoError(t, err)
	testCases := []struct {
		name      string
		i         *Inserter[TestModel]
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "single row",
			i:    NewInserter[TestModel](db).Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`id`,`first_name`,`last_name`) VALUES (?,?,?);",
				Args: []any{int64(1), "a", "b"},
			},
		},
		{
			name: "multi row",
			i: NewInserter[TestModel](db).Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"},
				&TestModel{Id: 2, FirstName: "c", LastName: "d"}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`id`,`first_name`,`last_name`) VALUES (?,?,?),(?,?,?);",
				Args: []any{int64(1), "a", "b", int64(2), "c", "d"},
			},
		},
		{
			name:    "no row",
			i:       NewInserter[TestModel](db),
			wantErr: errs.ErrInsertZeroRow,
		},
		{
			name: "specified columns",
			i: NewInserter[TestModel](db).Columns("FirstName", "LastName").
				Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"}),
			wantQuery: &Query{
				SQL:  "INSERT INTO `test_model` (`first_name`,`last_name`) VALUES (?,?);",
				Args: []any{"a", "b"},
			},
		},
		{
			name: "unknown column",
			i: NewInserter[TestModel](db).Columns("Unknown").
				Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"}),
			wantErr: errs.NewErrUnknownField("Unknown"),
		},
		{
			name: "upsert specific columns with new value",
			i: NewInserter[TestModel](db).
				Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"}).
				OnDuplicateKey().Update(Assign("FirstName", "newA"), Assign("LastName", "newB")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`,`last_name`) VALUES (?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name` = ?,`last_name` = ?;",
				Args: []any{int64(1), "a", "b", "newA", "newB"},
			},
		},
		{
			name: "upsert specific columns with column value",
			i: NewInserter[TestModel](db).Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"}).
				OnDuplicateKey().Update(Col("FirstName"), Col("LastName")),
			wantQuery: &Query{
				SQL: "INSERT INTO `test_model` (`id`,`first_name`,`last_name`) VALUES (?,?,?) " +
					"ON DUPLICATE KEY UPDATE `first_name` = VALUES(`first_name`),`last_name` = VALUES(`last_name`);",
				Args: []any{int64(1), "a", "b"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.i.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestInserter_Exec(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB, DBWithDialect(DialectMySQL))
	testCases := []struct {
		name     string
		i        *Inserter[TestModel]
		wantErr  error
		affected int64
	}{
		{
			name: "db err",
			i: func() *Inserter[TestModel] {
				mock.ExpectExec("INSERT INTO `test_model` .*").
					WillReturnError(errors.New("db err"))
				return NewInserter[TestModel](db).Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"})
			}(),
			wantErr: errors.New("db err"),
		},
		{
			name: "sql build err",
			i: func() *Inserter[TestModel] {
				return NewInserter[TestModel](db).Columns("Unknown").
					Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"})
			}(),
			wantErr: errs.NewErrUnknownField("Unknown"),
		},
		{
			name: "insert one row",
			i: func() *Inserter[TestModel] {
				res := driver.RowsAffected(1)
				mock.ExpectExec("INSERT INTO `test_model` .*").
					WillReturnResult(res)
				return NewInserter[TestModel](db).Values(&TestModel{Id: 1, FirstName: "a", LastName: "b"})
			}(),
			affected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.i.Exec(context.Background())
			affected, err := res.RowsAffected()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.affected, affected)
		})
	}

}
