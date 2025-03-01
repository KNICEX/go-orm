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

func TestSelector_Join(t *testing.T) {
	db := memoryDB(t)
	type Order struct {
		Id        int
		UsingCol1 string
		UsingCol2 string
	}

	type OrderDetail struct {
		OrderId   int
		ItemId    int
		UsingCol1 string
		UsingCol2 string
	}

	type Item struct {
		Id int
	}
	testCases := []struct {
		name      string
		s         func() SqlBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "specify table",
			s: func() SqlBuilder {
				return NewSelector[Order](db).From(TableOf(&OrderDetail{}))
			},
			wantQuery: &Query{
				SQL: "SELECT * FROM `order_detail`;",
			},
		},
		{
			name: "join-using",
			s: func() SqlBuilder {
				return NewSelector[Order](db).From(TableOf(&Order{}).Join(TableOf(&OrderDetail{})).Using("UsingCol1", "UsingCol2"))
			},
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` INNER JOIN `order_detail` USING (`using_col1`,`using_col2`));",
			},
		},
		{
			name: "join-on",
			s: func() SqlBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				return NewSelector[Order](db).From(t1.Join(t2).On(t1.Col("Id").Eq(t2.Col("OrderId"))))
			},
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` AS `t1` INNER JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`);",
			},
		},
		{
			name: "left join",
			s: func() SqlBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				return NewSelector[Order](db).From(t1.LeftJoin(t2).On(t1.Col("Id").Eq(t2.Col("OrderId"))))
			},
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` AS `t1` LEFT JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`);",
			},
		},
		{
			name: "join join",
			s: func() SqlBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.Col("Id").Eq(t2.Col("OrderId")))
				t4 := TableOf(&Item{}).As("t4")
				t5 := t3.Join(t4).On(t2.Col("ItemId").Eq(t4.Col("Id")))
				return NewSelector[Order](db).From(t5)
			},
			wantQuery: &Query{
				SQL: "SELECT * FROM ((`order` AS `t1` INNER JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) INNER JOIN `item` AS `t4` ON `t2`.`item_id` = `t4`.`id`);",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.s().Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
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
			builder: NewSelector[TestModel](db).From(TableOf(&TestModel{})),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			name:    "from empty",
			builder: NewSelector[TestModel](db).From(nil),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		//{
		//	name:    "from table with db",
		//	builder: NewSelector[TestModel](db).From("db.test"),
		//	wantQuery: &Query{
		//		SQL: "SELECT * FROM db.test;",
		//	},
		//},
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
			name:    "order by name",
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

func TestSelector_Count(t *testing.T) {
	db, err := OpenDB(nil, DBWithDialect(DialectMySQL))
	require.NoError(t, err)
	testCases := []struct {
		name      string
		s         SqlBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "count",
			s:    NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL: "SELECT COUNT(*) FROM `test_model`;",
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

func TestSelector_SubQuery(t *testing.T) {
	db := memoryDB(t)

	type Order struct {
		Id        int
		UsingCol1 string
		UsingCol2 string
	}

	type OrderDetail struct {
		OrderId int
		ItemId  int

		UsingCol1 string
		UsingCol2 string
	}

	type Item struct {
		Id int
	}

	testCases := []struct {
		name      string
		s         func() SqlBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "sub query",
			s: func() SqlBuilder {
				sub := NewSelector[OrderDetail](db).AsSubQuery("sub")
				return NewSelector[Order](db).Select(sub.Col("ItemId")).From(sub)
			},
			wantQuery: &Query{
				SQL: "SELECT `sub`.`item_id` FROM (SELECT * FROM `order_detail`) AS `sub`;",
			},
		},
		{
			name: "sub query with alias",
			s: func() SqlBuilder {
				sub := NewSelector[OrderDetail](db).AsSubQuery("sub")
				return NewSelector[Order](db).Select(sub.Col("ItemId").As("item_id")).From(sub)
			},
			wantQuery: &Query{
				SQL: "SELECT `sub`.`item_id` AS `item_id` FROM (SELECT * FROM `order_detail`) AS `sub`;",
			},
		},
		{
			name: "sub specify columns",
			s: func() SqlBuilder {
				sub := NewSelector[OrderDetail](db).Select(Col("OrderId"), Col("ItemId")).AsSubQuery("sub")
				return NewSelector[Order](db).Select(sub.Col("ItemId")).From(sub)
			},
			wantQuery: &Query{
				SQL: "SELECT `sub`.`item_id` FROM (SELECT `order_id`,`item_id` FROM `order_detail`) AS `sub`;",
			},
		},
		{
			name: "sub query with join",
			s: func() SqlBuilder {
				sub := NewSelector[OrderDetail](db).AsSubQuery("sub")
				return NewSelector[Order](db).Select(sub.Col("ItemId")).From(sub.Join(TableOf(&Item{})).On(sub.Col("ItemId").Eq(Col("Id"))))
			},
			wantQuery: &Query{
				SQL: "SELECT `sub`.`item_id` FROM ((SELECT * FROM `order_detail`) AS `sub` INNER JOIN `item` ON `sub`.`item_id` = `id`);",
			},
		},
		{
			name: "sub join sub",
			s: func() SqlBuilder {
				sub1 := NewSelector[OrderDetail](db).AsSubQuery("sub1")
				sub2 := NewSelector[Item](db).AsSubQuery("sub2")
				return NewSelector[Order](db).Select(sub1.Col("ItemId")).From(sub1.LeftJoin(sub2).On(sub1.Col("ItemId").Eq(sub2.Col("Id"))))

			},
			wantQuery: &Query{
				SQL: "SELECT `sub1`.`item_id` FROM ((SELECT * FROM `order_detail`) AS `sub1` LEFT JOIN (SELECT * FROM `item`) AS `sub2` ON `sub1`.`item_id` = `sub2`.`id`);",
			},
		},
		{
			name: "sub with aggregate",
			s: func() SqlBuilder {
				sub := NewSelector[OrderDetail](db).AsSubQuery("sub")
				return NewSelector[Order](db).Select(sub.Col("OrderId"), sub.Min("ItemId").As("min_id")).From(sub).GroupBy(sub.Col("OrderId"))
			},
			wantQuery: &Query{
				SQL: "SELECT `sub`.`order_id`,MIN(`sub`.`item_id`) AS `min_id` FROM (SELECT * FROM `order_detail`) AS `sub` GROUP BY `sub`.`order_id`;",
			},
		},
		{
			name: "sub join group having aggregate",
			s: func() SqlBuilder {
				sub1 := NewSelector[OrderDetail](db).Select(Col("OrderId"), Col("ItemId")).AsSubQuery("sub1")
				sub2 := NewSelector[Item](db).AsSubQuery("sub2")
				return NewSelector[Order](db).From(sub1.Join(sub2).On(sub1.Col("ItemId").Eq(sub2.Col("Id")))).GroupBy(sub1.Col("OrderId")).Having(sub1.Min("ItemId").Gt(1))
			},
			wantQuery: &Query{
				SQL:  "SELECT * FROM ((SELECT `order_id`,`item_id` FROM `order_detail`) AS `sub1` INNER JOIN (SELECT * FROM `item`) AS `sub2` ON `sub1`.`item_id` = `sub2`.`id`) GROUP BY `sub1`.`order_id` HAVING MIN(`sub1`.`item_id`) > ?;",
				Args: []any{1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := tc.s().Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, q)
		})
	}
}
