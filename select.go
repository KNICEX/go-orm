package orm

import (
	"context"
	"github.com/KNICEX/go-orm/internal/errs"
)

// Selectable 标记接口
type Selectable interface {
	selectable()
}

type GroupAble interface {
	groupAble()
}
type OrderAble interface {
	orderAble()
}

type Selector[T any] struct {
	table    string
	where    []Predicate
	columns  []Selectable
	orderBys []OrderAble
	groupBys []GroupAble
	having   []Predicate
	// 只查询COUNT
	count  bool
	offset int
	limit  int

	builder
	sess Session
}

func NewSelector[T any](sess Session) *Selector[T] {
	c := sess.getCore()
	return &Selector[T]{
		sess: sess,
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
	}
}

func (s *Selector[T]) Build() (*Query, error) {
	m, err := s.r.Register(new(T))
	if err != nil {
		return nil, err
	}
	s.model = m

	s.sb.WriteString("SELECT ")

	if s.count {
		// 只查询 COUNT
		s.sb.WriteString("COUNT(*)")
	} else if err = s.buildColumns(s.columns); err != nil {
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

	// 排序
	if len(s.orderBys) > 0 {
		s.sb.WriteString(" ORDER BY ")
		for i, ob := range s.orderBys {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			switch o := ob.(type) {
			case Column:
				fd, ok := s.model.FieldMap[o.name]
				if !ok {
					return nil, errs.NewErrUnknownField(o.name)
				}
				s.quote(fd.ColName)
				if o.desc {
					s.sb.WriteString(" DESC")
				} else {
					s.sb.WriteString(" ASC")
				}
			case RawExpr:
				s.sb.WriteByte('(')
				s.sb.WriteString(o.raw)
				s.sb.WriteByte(')')
			}
		}
	}

	// 分组
	if len(s.groupBys) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, gb := range s.groupBys {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			switch g := gb.(type) {
			case Column:
				fd, ok := s.model.FieldMap[g.name]
				if !ok {
					return nil, errs.NewErrUnknownField(g.name)
				}
				s.quote(fd.ColName)
			case RawExpr:
				s.sb.WriteByte('(')
				s.sb.WriteString(g.raw)
				s.sb.WriteByte(')')
			}
		}

		if len(s.having) > 0 {
			// having
			s.sb.WriteString(" HAVING ")
			if err = s.buildPredicate(s.having); err != nil {
				return nil, err
			}
		}

	}

	// limit offset
	if err = s.dialect.buildOffsetLimit(&s.builder, s.offset, s.limit); err != nil {
		return nil, err
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

func getHandler(ctx *Context, sess Session, c *core, entity any) *Result {
	rows, err := sess.queryContext(ctx.Ctx, ctx.Query.SQL, ctx.Query.Args...)
	if err != nil {
		return &Result{Err: err}
	}
	defer rows.Close()
	if !rows.Next() {
		return &Result{Err: ErrNoRows}
	}

	val := c.creator(c.model, entity)
	err = val.SetColumns(rows)
	if err != nil {
		return &Result{
			Err: err,
		}
	}
	return &Result{
		Res: entity,
	}

}

func get(ctx context.Context, builder SqlBuilder, sess Session, c *core, opType string, entity any) error {
	q, err := builder.Build()
	if err != nil {
		return err
	}

	var root Handler = func(ctx *Context) *Result {
		return getHandler(ctx, sess, c, entity)
	}

	for i := len(c.middlewares) - 1; i >= 0; i-- {
		root = c.middlewares[i](root)
	}

	res := root(&Context{
		Type:  opType,
		Query: q,
		Model: c.model,
		Ctx:   ctx,
	})

	if res.Err != nil {
		return res.Err
	}
	return nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	s.limit = 1
	resEntity := new(T)
	err := get(ctx, s, s.sess, s.core, SELECT, resEntity)
	if err != nil {
		return nil, err
	}
	return resEntity, nil
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	resEntity := new([]*T)
	err := get(ctx, s, s.sess, s.core, SELECT, resEntity)
	if err != nil {
		return nil, err
	}
	return *resEntity, nil

}

func (s *Selector[T]) countHandler(ctx *Context) *Result {
	rows, err := s.sess.queryContext(ctx.Ctx, ctx.Query.SQL, ctx.Query.Args...)
	if err != nil {
		return &Result{
			Err: err,
		}
	}
	defer rows.Close()

	if !rows.Next() {
		return &Result{
			Err: ErrNoRows,
		}
	}
	var count int64
	err = rows.Scan(&count)
	return &Result{
		Res: count,
		Err: err,
	}
}

func (s *Selector[T]) Count(ctx context.Context) (int64, error) {
	s.count = true
	q, err := s.Build()
	if err != nil {
		return 0, err
	}

	root := s.countHandler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		root = s.middlewares[i](root)
	}
	res := root(&Context{
		Type:  SELECT,
		Query: q,
		Model: s.model,
		Ctx:   ctx,
	})
	if res.Res != nil {
		return res.Res.(int64), nil
	}
	return 0, res.Err
}

func (s *Selector[T]) Where(p Predicate) *Selector[T] {
	s.where = append(s.where, p)
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) OrderBy(orderBys ...OrderAble) *Selector[T] {
	s.orderBys = orderBys
	return s
}

func (s *Selector[T]) GroupBy(groupBys ...GroupAble) *Selector[T] {
	s.groupBys = groupBys
	return s
}

func (s *Selector[T]) Having(p Predicate) *Selector[T] {
	s.having = append(s.having, p)
	return s
}
