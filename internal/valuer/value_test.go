package valuer

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KNICEX/go-orm/model"
	"github.com/stretchr/testify/require"
	"testing"
)

// go test -bench=BenchmarkSetColumns -benchmem -benchtime=10000x
func BenchmarkSetColumns(b *testing.B) {

	fn := func(b *testing.B, creator Creator) {
		mockRows := sqlmock.NewRows([]string{"id", "last_name", "first_name"})
		row := []driver.Value{1, "Smith", "John"}
		for i := 0; i < b.N; i++ {
			mockRows.AddRow(row...)
		}
		mockDB, mock, err := sqlmock.New()
		require.NoError(b, err)
		mock.ExpectQuery("SELECT .*").WillReturnRows(mockRows)
		rows, err := mockDB.Query("SELECT XX")
		require.NoError(b, err)
		r := model.NewRegistry()
		m, err := r.Get(&TestModel{})
		require.NoError(b, err)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rows.Next()
			val := creator(m, &TestModel{})
			_ = val.SetColumns(rows)
		}
	}

	b.Run("reflect", func(b *testing.B) {
		fn(b, NewReflectValue)
	})

	b.Run("unsafe", func(b *testing.B) {
		fn(b, NewUnsafeValue)
	})
}
