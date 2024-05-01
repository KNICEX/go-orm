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

type TestModel struct {
	Id        int64
	FirstName string
	LastName  string
}

func TestSelector_Build(t *testing.T) {
	db, err := OpenDB(nil)
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
		{
			name:    "raw where",
			builder: NewSelector[TestModel](db).Where(Raw("id = ? AND first_name = ?", 18, "tom").AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (id = ? AND first_name = ?);",
				Args: []any{18, "tom"},
			},
		},
		{
			name:    "raw in predicate",
			builder: NewSelector[TestModel](db).Where(Col("Id").Eq(Raw("age + ?", 1))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = (age + ?);",
				Args: []any{1},
			},
		},
		{
			// 条件表达式中别名会被忽略
			name:    "alias in predicate",
			builder: NewSelector[TestModel](db).Where(Col("Id").As("user_id").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = ?;",
				Args: []any{18},
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

// 必须一次运行所有测试用例，因为 mock.ExpectQuery() 会按照调用顺序匹配
func TestSelector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
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
		s       *Selector[TestModel]
		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "invalid query",
			s:       NewSelector[TestModel](db).Where(Col("xx").Eq(18)),
			wantErr: errs.NewErrUnknownField("xx"),
		},
		{
			name:    "query error",
			s:       NewSelector[TestModel](db).Where(Col("Id").Eq(1)),
			wantErr: errors.New("query error"),
		},
		{
			name:    "no rows",
			s:       NewSelector[TestModel](db).Where(Col("Id").Ge(1)),
			wantErr: errs.ErrNoRows,
		},
		{
			name: "data",
			s:    NewSelector[TestModel](db).Where(Col("Id").Eq(1)),
			wantRes: &TestModel{
				Id:        1,
				FirstName: "tom",
				LastName:  "cat",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestSelector_GetMulti(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db, err := OpenDB(mockDB)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		s       *Selector[TestModel]
		rows    *sqlmock.Rows
		wantErr error
		wantRes []*TestModel
	}{
		{
			name: "success",
			s:    NewSelector[TestModel](db),
			rows: sqlmock.NewRows([]string{"id", "first_name", "last_name"}).
				AddRow(1, "tom", "cat").
				AddRow(2, "jerry", "mouse"),
			wantRes: []*TestModel{
				{
					Id:        1,
					FirstName: "tom",
					LastName:  "cat",
				},
				{
					Id:        2,
					FirstName: "jerry",
					LastName:  "mouse",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectQuery("SELECT .*").WillReturnRows(tc.rows)
			res, err := tc.s.GetMulti(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestSelector_Select(t *testing.T) {
	db, err := OpenDB(nil)
	require.NoError(t, err)
	testCases := []struct {
		name      string
		s         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "multiple columns",
			s:    NewSelector[TestModel](db).Select(Col("Id"), Col("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT `id`,`first_name` FROM `test_model`;",
			},
		},
		{
			name: "columns with alias",
			s:    NewSelector[TestModel](db).Select(Col("Id").As("user_id"), Col("FirstName").As("name")),
			wantQuery: &Query{
				SQL: "SELECT `id` AS `user_id`,`first_name` AS `name` FROM `test_model`;",
			},
		},
		{
			name:    "error column name",
			s:       NewSelector[TestModel](db).Select(Col("xx")),
			wantErr: errs.NewErrUnknownField("xx"),
		},
		{
			name: "aggregate AVG, COUNT",
			s:    NewSelector[TestModel](db).Select(Avg("Id"), Count("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`id`),COUNT(`first_name`) FROM `test_model`;",
			},
		},
		{
			name: "aggregate with alias",
			s:    NewSelector[TestModel](db).Select(Avg("Id").As("avg_id"), Count("FirstName").As("count_name")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`id`) AS `avg_id`,COUNT(`first_name`) AS `count_name` FROM `test_model`;",
			},
		},
		{
			name:    "invalid  Min column",
			s:       NewSelector[TestModel](db).Select(Min("xx")),
			wantErr: errs.NewErrUnknownField("xx"),
		},

		{
			name: "raw expression",
			s:    NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT first_name)")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(DISTINCT first_name) FROM `test_model`;",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.s.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}
