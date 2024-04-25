package orm

import (
	"context"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	table string
	where []Predicate
	sb    strings.Builder
	args  []any
}

func (s *Selector[T]) Build() (*Query, error) {
	sb := &s.sb
	sb.WriteString("SELECT * FROM ")

	// 表名 如果没有指定表名，则使用类型名
	if s.table == "" {
		var t T
		sb.WriteByte('`')
		sb.WriteString(reflect.TypeOf(t).Name())
		sb.WriteByte('`')
	} else {
		// 自己指定表名，不会自动加反引号， 因为可能是 db.table 这种形式
		sb.WriteString(s.table)
	}

	// 条件构造
	if len(s.where) > 0 {
		sb.WriteString(" WHERE ")
		p := s.where[0]
		// 如果使用了多个Where条件，需要用 AND 连接
		for i := 1; i < len(s.where); i++ {
			p = p.And(s.where[i])
		}

		if err := s.buildExpression(p); err != nil {
			return nil, err
		}

	}

	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: s.args,
	}, nil
}

// buildExpression 递归构造表达式
func (s *Selector[T]) buildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case Predicate:

		// 如果左边也是一个表达式，那么需要加括号
		_, ok := exp.left.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

		// 操作符
		s.sb.WriteByte(' ')
		s.sb.WriteString(exp.op.String())
		s.sb.WriteByte(' ')

		// 如果右边也是一个表达式，那么需要加括号
		_, ok = exp.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')

		}
	case Column:
		s.sb.WriteByte('`')
		s.sb.WriteString(exp.name)
		s.sb.WriteByte('`')
	case value:
		s.sb.WriteByte('?')
		s.args = append(s.args, exp.val)
	}
	return nil
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
