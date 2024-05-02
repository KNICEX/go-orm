package orm

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeleter_Build(t *testing.T) {
	db, err := OpenDB(nil, DBWithDialect(DialectMySQL))
	require.NoError(t, err)
	testCases := []struct {
		name      string
		d         *Deleter[TestModel]
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "test_model",
			d:    NewDeleter[TestModel](db),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
		},
		{
			name: "where",
			d:    NewDeleter[TestModel](db).Where(Col("Id").Eq(18)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE `id` = ?;",
				Args: []any{18},
			},
		},
		{
			name: "multiple where",
			d:    NewDeleter[TestModel](db).Where(Col("Id").Eq(18).Or(Col("Id").Eq(19))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE (`id` = ?) OR (`id` = ?);",
				Args: []any{18, 19},
			},
		},
		{
			name: "raw where",
			d:    NewDeleter[TestModel](db).Where(Raw("id = ?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "DELETE FROM `test_model` WHERE (id = ?);",
				Args: []any{18},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.d.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}
