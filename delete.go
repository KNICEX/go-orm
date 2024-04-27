package orm

import (
	"context"
	"database/sql"
)

type Deleter[T any] struct {
	table string
	where []Predicate

	builder

	model *model
	r     *registry
}

func (d *Deleter[T]) Build() (*Query, error) {
	m, err := d.r.get(new(T))
	if err != nil {
		return nil, err
	}
	d.model = m
	sb := &d.sb
	sb.WriteString("DELETE FROM ")

	// 表名 如果没有指定表名，则使用类型名
	if d.table == "" {
		sb.WriteByte('`')
		sb.WriteString(m.tableName)
		sb.WriteByte('`')
	} else {
		// 自己指定表名，不会自动加反引号， 因为可能是 db.table 这种形式
		sb.WriteString(d.table)
	}

	// 条件构造
	if len(d.where) > 0 {
		sb.WriteString(" WHERE ")

		if err := d.buildPredicate(d.where); err != nil {
			return nil, err
		}

	}

	sb.WriteByte(';')
	return &Query{
		SQL:  sb.String(),
		Args: d.args,
	}, nil
}

func (d *Deleter[T]) From(table string) *Deleter[T] {
	d.table = table
	return d
}

func (d *Deleter[T]) Where(p Predicate) *Deleter[T] {
	d.where = append(d.where, p)
	return d
}

func (d *Deleter[T]) Exec(ctx context.Context) (sql.Result, error) {
	panic("implement me")
}
