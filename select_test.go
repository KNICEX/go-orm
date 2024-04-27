package orm

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestModel struct {
	Id        int64
	LastName  string
	FirstName string
}

func TestSelector_Build(t *testing.T) {
	db, err := NewDB()
	require.NoError(t, err)
	testCases := []struct {
		name      string
		builder   QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "test_model",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:    "from",
			builder: NewSelector[TestModel](db).From("test"),
			wantQuery: &Query{
				SQL: "SELECT * FROM test;",
			},
		},
		{
			name:    "from empty",
			builder: NewSelector[TestModel](db).From(""),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:    "from table with db",
			builder: NewSelector[TestModel](db).From("db.test"),
			wantQuery: &Query{
				SQL: "SELECT * FROM db.test;",
			},
		},
		{
			name:    "where",
			builder: NewSelector[TestModel](db).Where(Col("Id").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = ?;",
				Args: []any{18},
			},
		},
		{
			name:    "not",
			builder: NewSelector[TestModel](db).Where(Not(Col("Id").Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`id` = ?);",
				Args: []any{18},
			},
		},
		{
			name:    "or",
			builder: NewSelector[TestModel](db).Where(Col("Id").Eq(18).Or(Col("LastName").Eq("hello"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`id` = ?) OR (`last_name` = ?);",
				Args: []any{18, "hello"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.builder.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}

func TestRotate(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6, 7}
	rotate(nums, 3)
	assert.Equal(t, []int{5, 6, 7, 1, 2, 3, 4}, nums)
}

func rotate(nums []int, k int) {
	n := len(nums)
	offset := k % n
	if offset == 0 {
		return
	}

	reverse(nums, 0, n)
	reverse(nums, 0, offset)
	reverse(nums, offset, n)

}

func reverse(nums []int, i, j int) {
	n := (i + j) / 2
	for start := i; start < n; start++ {
		nums[start], nums[j-1-(start-i)] = nums[j-1-(start-i)], nums[start]
	}
}
