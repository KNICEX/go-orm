package orm

import (
	"context"
)

type Selector[T any] struct {
	table string
	where []Predicate

	builder

	model *model
	db    *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	m, err := s.db.r.parseModel(new(T))
	if err != nil {
		return nil, err
	}
	s.model = m
	sb := &s.sb
	sb.WriteString("SELECT * FROM ")

	// 表名 如果没有指定表名，则使用类型名
	if s.table == "" {
		sb.WriteByte('`')
		sb.WriteString(m.tableName)
		sb.WriteByte('`')
	} else {
		// 自己指定表名，不会自动加反引号， 因为可能是 db.table 这种形式
		sb.WriteString(s.table)
	}

	// 条件构造
	if len(s.where) > 0 {
		sb.WriteString(" WHERE ")

		if err := s.buildPredicate(s.where); err != nil {
			return nil, err
		}

	}

	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) Where(p Predicate) *Selector[T] {
	s.where = append(s.where, p)
	return s
}

// Selector[User].Eq("age", 18).Eq("name", "foo")

// Selector[User].Where(Col("age").Eq(18).And(Col("name").Eq("foo")))
