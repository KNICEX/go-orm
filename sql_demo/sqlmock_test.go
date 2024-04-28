package sql_demo

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSQLMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	mockRows := sqlmock.NewRows([]string{"id", "first_name", "last_name"})
	mockRows.AddRow(1, "foo", "bar")
	mock.ExpectQuery("SELECT id, first_name, last_name FROM user").WillReturnRows(mockRows)

	rows, err := db.QueryContext(context.Background(), "SELECT id, first_name, last_name FROM user")
	require.NoError(t, err)
	for rows.Next() {
		tm := TestModel{}
		err = rows.Scan(&tm.Id, &tm.FirstName, &tm.LastName)
		require.NoError(t, err)
		require.Equal(t, TestModel{Id: 1, FirstName: "foo", LastName: "bar"}, tm)
	}
}
