package orm

import (
	"context"
)

type Deleter[T any] struct {
	table string
	where []Predicate

	builder

	db *DB
}

var _ Executor = (*Deleter[any])(nil)

func NewDeleter[T any](db *DB) *Deleter[T] {
	return &Deleter[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
	}
}

func (d *Deleter[T]) Build() (*Query, error) {
	m, err := d.db.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	d.model = m

	d.sb.WriteString("DELETE FROM ")
	// 表名 如果没有指定表名，则使用类型名
	if d.table == "" {
		d.quote(m.TableName)
	} else {
		// 自己指定表名，不会自动加反引号， 因为可能是 db.table 这种形式
		d.sb.WriteString(d.table)
	}

	// 条件构造
	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")

		if err := d.buildPredicate(d.where); err != nil {
			return nil, err
		}

	}

	d.sb.WriteByte(';')
	return &Query{
		SQL:  d.sb.String(),
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

func (d *Deleter[T]) Exec(ctx context.Context) Result {
	q, err := d.Build()
	if err != nil {
		return Result{
			err: err,
		}
	}
	res, err := d.db.db.ExecContext(ctx, q.SQL, q.Args...)
	return Result{
		res: res,
		err: err,
	}
}
