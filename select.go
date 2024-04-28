package orm

import (
	"context"
	"github.com/KNICEX/go-orm/internal/errs"
	model2 "github.com/KNICEX/go-orm/model"
	"reflect"
	"unsafe"
)

type Selector[T any] struct {
	table string
	where []Predicate

	builder

	model *model2.Model
	db    *DB
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	m, err := s.db.r.Register(new(T))
	if err != nil {
		return nil, err
	}
	s.model = m
	s.builder.model = m

	sb := &s.sb
	sb.WriteString("SELECT * FROM ")

	// 表名 如果没有指定表名，则使用类型名
	if s.table == "" {
		sb.WriteByte('`')
		sb.WriteString(m.TableName)
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

func (s *Selector[T]) GetV1(ctx context.Context) (*T, error) {
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

	cs, err := row.Columns()

	tp := new(T)
	vals := make([]any, 0, len(cs))

	//var creator valuer.Creator
	//err =  creator(tp).SetColumns(row)
	//return tp, err

	// 起始地址
	addr := reflect.ValueOf(tp).UnsafePointer()
	for _, c := range cs {
		fd, ok := s.model.ColMap[c]
		if !ok {
			return nil, errs.NewErrUnknownColumn(c)
		}
		// 字段地址
		fdAddr := unsafe.Pointer(uintptr(addr) + fd.Offset)
		// 指针类型
		val := reflect.NewAt(fd.Typ, fdAddr).Interface()
		vals = append(vals, val)
	}

	err = row.Scan(vals...)
	if err != nil {
		return nil, err
	}

	return tp, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
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

	tp := new(T)
	val := s.db.creator(s.model, tp)

	err = val.SetColumns(row)
	if err != nil {
		return nil, err
	}

	return tp, nil
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
