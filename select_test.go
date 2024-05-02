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
	db, err := OpenDB(nil, DBWithDialect(DialectMySQL))
	require.NoError(t, err)
	testCases := []struct {
		name      string
		builder   SqlBuilder
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

		{
			name:    "limit",
			builder: NewSelector[TestModel](db).Where(Col("Id").Ge(1)).Limit(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` >= ? LIMIT 10;",
				Args: []any{1},
			},
		},
		{
			name:    "offset",
			builder: NewSelector[TestModel](db).Where(Col("Id").Ge(1)).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` >= ? OFFSET 10;",
				Args: []any{1},
			},
		},
		{
			name:    "limit offset",
			builder: NewSelector[TestModel](db).Where(Col("Id").Ge(1)).Limit(10).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` >= ? LIMIT 10 OFFSET 10;",
				Args: []any{1},
			},
		},

		{
			name:    "order by col",
			builder: NewSelector[TestModel](db).OrderBy(Col("Id").Desc(), Col("FirstName").Asc()),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` ORDER BY `id` DESC,`first_name` ASC;",
			},
		},
		{
			name:    "order by raw",
			builder: NewSelector[TestModel](db).OrderBy(Raw("RAND()")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` ORDER BY (RAND());",
			},
		},

		{
			name:    "group by",
			builder: NewSelector[TestModel](db).Select(Col("FirstName"), Sum("FirstName")).GroupBy(Col("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT `first_name`,SUM(`first_name`) FROM `test_model` GROUP BY `first_name`;",
			},
		},

		{
			name: "group by with having",
			builder: NewSelector[TestModel](db).Select(Col("FirstName"), Sum("FirstName")).
				GroupBy(Col("FirstName")).
				Having(Sum("FirstName").Gt(1)),
			wantQuery: &Query{
				SQL: "SELECT `first_name`,SUM(`first_name`) FROM `test_model` " +
					"GROUP BY `first_name` HAVING SUM(`first_name`) > ?;",
				Args: []any{1},
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

// 必须一次运行所有测试用例，因为 mock.ExpectQuery() 会按照调用顺序匹配
func TestSelector_Get(t *testing.T) {
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
	db, err := OpenDB(mockDB, DBWithDialect(DialectMySQL))
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
	db, err := OpenDB(nil, DBWithDialect(DialectMySQL))
	require.NoError(t, err)
	testCases := []struct {
		name      string
		s         SqlBuilder
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
