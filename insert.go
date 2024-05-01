package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/KNICEX/go-orm/model"
	"reflect"
)

type OnDuplicateKeyBuilder[T any] struct {
	i *Inserter[T]
}

type OnDuplicateKey struct {
	assigns []Assignable
}

func (o *OnDuplicateKeyBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &OnDuplicateKey{
		assigns: assigns,
	}
	return o.i
}

type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	db *DB
	builder

	values  []*T
	columns []string

	onDuplicateKey *OnDuplicateKey
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
	}
}

func (i *Inserter[T]) OnDuplicateKey() *OnDuplicateKeyBuilder[T] {
	return &OnDuplicateKeyBuilder[T]{
		i: i,
	}
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}

	i.sb.WriteString("INSERT INTO ")
	m, err := i.db.r.Get(i.values[0])
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

		refVal := reflect.ValueOf(v).Elem()
		for j, field := range fields {
			if j > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			i.addArgs(refVal.FieldByName(field.GoName).Interface())
		}

		i.sb.WriteByte(')')
	}

	// 处理 ON DUPLICATE KEY UPDATE
	if i.onDuplicateKey != nil {
		err = i.dialect.buildOnDuplicateKey(&i.builder, i.onDuplicateKey)
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
