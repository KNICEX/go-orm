package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInserter_Build(t *testing.T) {
	db, err := OpenDB(nil)
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
