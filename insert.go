package orm

import (
	"context"
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/KNICEX/go-orm/model"
)

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

func (o *UpsertBuilder[T]) ConflictColumns(columns ...string) *UpsertBuilder[T] {
	o.conflictColumns = columns
	return o
}

type Assignable interface {
	assign()
}

var _ Executor = (*Inserter[any])(nil)

type Inserter[T any] struct {
	sess Session
	builder

	values  []*T
	columns []string

	onDuplicateKey *Upsert
}

func NewInserter[T any](sess Session) *Inserter[T] {
	c := sess.getCore()
	return &Inserter[T]{
		sess: sess,
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
	}
}

func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}

	i.sb.WriteString("INSERT INTO ")
	m, err := i.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	i.model = m

	i.quote(m.TableName)
	i.sb.WriteByte(' ')

	// 构造列名
	i.sb.WriteByte('(')

	fields := m.Fields

	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, fd := range i.columns {
			fdMeta, ok := m.FieldMap[fd]
			if !ok {
				return nil, errs.NewErrUnknownField(fd)
			}
			fields = append(fields, fdMeta)
		}
	}

	for idx, field := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(field.ColName)
	}
	i.sb.WriteByte(')')

	// 构造值
	i.sb.WriteString(" VALUES ")

	i.args = make([]any, 0, len(fields)*len(i.values))
	for idx, v := range i.values {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		val := i.creator(m, v)
		for j, field := range fields {
			if j > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			arg, err := val.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArgs(arg)
		}

		i.sb.WriteByte(')')
	}

	// 处理 ON DUPLICATE KEY UPDATE
	if i.onDuplicateKey != nil {
		err = i.dialect.buildUpsert(&i.builder, i.onDuplicateKey)
		if err != nil {
			return nil, err
		}
	}

	i.sb.WriteByte(';')

	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

func (i *Inserter[T]) Values(values ...*T) *Inserter[T] {
	i.values = append(i.values, values...)
	return i
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Exec(ctx context.Context) Result {
	query, err := i.Build()
	if err != nil {
		return Result{err: err}
	}
	res, err := i.sess.execContext(ctx, query.SQL, query.Args...)
	return Result{
		res: res,
		err: err,
	}
}
