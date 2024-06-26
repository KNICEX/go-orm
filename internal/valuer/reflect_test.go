package valuer

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KNICEX/go-orm/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestModel struct {
	Id        int64
	LastName  string
	FirstName string
}

func TestReflectValue_SetColumns(t *testing.T) {
	testCases := []struct {
		name string
		// 结构体指针
		entity     any
		rows       func() *sqlmock.Rows
		wantErr    error
		wantEntity any
	}{
		{
			name:   "success",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "last_name", "first_name"}).
					AddRow(1, "Smith", "John")
				return rows
			},
			wantEntity: &TestModel{
				Id:        1,
				LastName:  "Smith",
				FirstName: "John",
			},
		},
		{
			name:   "slice",
			entity: &[]*TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "last_name", "first_name"}).
					AddRow(1, "Smith", "John").
					AddRow(2, "Doe", "Jane")
				return rows
			},
			wantEntity: &[]*TestModel{
				{
					Id:        1,
					LastName:  "Smith",
					FirstName: "John",
				},
				{
					Id:        2,
					LastName:  "Doe",
					FirstName: "Jane",
				},
			},
		},
	}

	r := model.NewRegistry()
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model, err := r.Get(&TestModel{})
			require.NoError(t, err)
			val := NewReflectValue(model, tc.entity)

			// 将sqlmock.Rows转为*sql.Rows
			mockRows := tc.rows()
			mock.ExpectQuery("SELECT .*").WillReturnRows(mockRows)
			rows, err := mockDB.Query("SELECT XX")
			require.NoError(t, err)
			rows.Next()

			err = val.SetColumns(rows)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.EqualValues(t, tc.wantEntity, tc.entity)

		})
	}

}
