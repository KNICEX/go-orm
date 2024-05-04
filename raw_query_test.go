package orm

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRawQuery_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB, DBWithDialect(DialectMySQL))
	require.NoError(t, err)

	// 查询错误
	mock.ExpectQuery("SELECT .*").WillReturnError(errors.New("query error"))

	// 查询结果为空
	rows := sqlmock.NewRows([]string{"id", "first_name", "last_name"})
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	// 查询结果
	rows = sqlmock.NewRows([]string{"id", "first_name", "last_name"})
	rows.AddRow("1", "tom", "cat")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	testCases := []struct {
		name    string
		r       *RawQuerier[TestModel]
		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "query error",
			r:       RawQuery[TestModel](db, "SELECT * FROM test_model;"),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			r:       RawQuery[TestModel](db, "SELECT * FROM test_model;"),
			wantErr: errs.ErrNoRows,
		},
		{
			name: "data",
			r:    RawQuery[TestModel](db, "SELECT * FROM test_model WHERE id = ?;", 1),
			wantRes: &TestModel{
				Id:        1,
				FirstName: "tom",
				LastName:  "cat",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.r.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
