package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdater_Build(t *testing.T) {
	db, err := OpenDB(nil, DBWithDialect(DialectMySQL))
	assert.NoError(t, err)
	testCases := []struct {
		name      string
		u         *Updater[TestModel]
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "no set",
			u:       NewUpdater[TestModel](db),
			wantErr: errs.ErrUpdateNoSet,
		},
		{
			name: "set assign",
			u: NewUpdater[TestModel](db).
				Set(Assign("FirstName", "newA"), Assign("LastName", "newB")),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `first_name` = ?,`last_name` = ?;",
				Args: []any{"newA", "newB"},
			},
		},
		{
			name: "set raw",
			u: NewUpdater[TestModel](db).
				Set(Raw("first_name = ?", "newA"), Raw("last_name = first_name")),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET first_name = ?,last_name = first_name;",
				Args: []any{"newA"},
			},
		},

		{
			name: "with where",
			u: NewUpdater[TestModel](db).
				Set(Assign("FirstName", "newA"), Assign("LastName", "newB")).
				Where(Col("Id").Eq(1)),
			wantQuery: &Query{
				SQL:  "UPDATE `test_model` SET `first_name` = ?,`last_name` = ? WHERE `id` = ?;",
				Args: []any{"newA", "newB", 1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.u.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}
