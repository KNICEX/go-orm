package orm

import (
	"context"
	"database/sql"
)

// Selectable 标记接口
type Selectable interface {
	selectable()
}

type Selector[T any] struct {
	table   string
	where   []Predicate
	columns []Selectable

	builder

	db *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	m, err := s.db.r.Register(new(T))
	if err != nil {
		return nil, err
	}
	s.model = m

	s.sb.WriteString("SELECT ")

	if err = s.buildColumns(s.columns); err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")
	// 表名 如果没有指定表名，则使用类型名
	if s.table == "" {
		s.quote(m.TableName)
	} else {
		// 自己指定表名，不会自动加反引号， 因为可能是 db.table 这种形式
		s.sb.WriteString(s.table)
	}

	// 条件构造
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")

		if err := s.buildPredicate(s.where); err != nil {
			return nil, err
		}

	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

// beforeQuery 查询和查询前后的检查
func (s *Selector[T]) beforeQuery(ctx context.Context) (*sql.Rows, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	db := s.db.db
	row, err := db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}

	if !row.Next() {
		return nil, ErrorNoRows
	}
	return row, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	row, err := s.beforeQuery(ctx)
	if err != nil {
		return nil, err
	}
	tp := new(T)
	val := s.db.creator(s.model, tp)

	err = val.SetColumns(row)
	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	rows, err := s.beforeQuery(ctx)
	if err != nil {
		return nil, err
	}
	var result []*T
	val := s.db.creator(s.model, &result)
	err = val.SetColumns(rows)
	return result, err
}

func (s *Selector[T]) Where(p Predicate) *Selector[T] {
	s.where = append(s.where, p)
	return s
}
