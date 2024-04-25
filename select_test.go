package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestModel struct {
	Id       int64
	LastName *sql.NullString
}

func TestSelector_Build(t *testing.T) {
	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "test_model",
			builder: &Selector[TestModel]{},
			wantQuery: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
		},
		{
			name:    "from",
			builder: (&Selector[TestModel]{}).From("test"),
			wantQuery: &Query{
				SQL: "SELECT * FROM test;",
			},
		},
		{
			name:    "from empty",
			builder: (&Selector[TestModel]{}).From(""),
			wantQuery: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
		},
		{
			name:    "from table with db",
			builder: (&Selector[TestModel]{}).From("db.test"),
			wantQuery: &Query{
				SQL: "SELECT * FROM db.test;",
			},
		},
		{
			name:    "where",
			builder: (&Selector[TestModel]{}).Where(Col("age").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE `age` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: (&Selector[TestModel]{}).Where(Not(Col("age").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "or",
			builder: (&Selector[TestModel]{}).Where(Col("age").Eq(18).Or(Col("age").Eq(20))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`age` = ?) OR (`age` = ?);",
				Args: []any{18, 20},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantQuery, q)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func aaa() {

}
