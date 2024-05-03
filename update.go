package orm

import (
	"context"
	"github.com/KNICEX/go-orm/internal/errs"
)

type SetAble interface {
	setAble()
}

type Updater[T any] struct {
	table string
	set   []SetAble
	where []Predicate

	builder
	sess Session
}

func NewUpdater[T any](sess Session) *Updater[T] {
	c := sess.getCore()
	return &Updater[T]{
		sess: sess,
		builder: builder{
			core:   c,
			quoter: c.dialect.quoter(),
		},
	}
}

func (u *Updater[T]) Build() (*Query, error) {
	m, err := u.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	u.model = m

	u.sb.WriteString("UPDATE ")
	if u.table == "" {
		u.quote(m.TableName)
	} else {
		u.sb.WriteString(u.table)
	}

	u.sb.WriteString(" SET ")
	if len(u.set) == 0 {
		return nil, errs.ErrUpdateNoSet
	}

	for i, s := range u.set {
		if i > 0 {
			u.sb.WriteByte(',')
		}
		switch v := s.(type) {
		case Assignment:
			fd, ok := m.FieldMap[v.col]
			if !ok {
				return nil, errs.NewErrUnknownField(v.col)
			}
			u.quote(fd.ColName)
			u.sb.WriteString(" = ?")
			u.addArgs(v.val)
		case RawExpr:
			u.sb.WriteString(v.raw)
			u.addArgs(v.args...)
		default:
			return nil, errs.NewErrUnsupportedSetAble(s)
		}
	}

	if len(u.where) > 0 {
		u.sb.WriteString(" WHERE ")
		if err = u.buildPredicate(u.where); err != nil {
			return nil, err
		}
	}

	u.sb.WriteByte(';')
	return &Query{
		SQL:  u.sb.String(),
		Args: u.args,
	}, nil

}

func (u *Updater[T]) Set(assignments ...SetAble) *Updater[T] {
	u.set = append(u.set, assignments...)
	return u
}

func (u *Updater[T]) Where(predicates Predicate) *Updater[T] {
	u.where = append(u.where, predicates)
	return u
}

func (u *Updater[T]) execHandler(ctx *Context) *Result {
	res, err := u.sess.execContext(ctx.Ctx, ctx.Query.SQL, ctx.Query.Args...)
	return &Result{
		Res: ExecResult{
			res: res,
			err: err,
		},
		Err: err,
	}
}

func (u *Updater[T]) Exec(ctx context.Context) ExecResult {
	q, err := u.Build()
	if err != nil {
		return ExecResult{
			err: err,
		}
	}

	root := u.execHandler
	for i := len(u.middlewares) - 1; i >= 0; i-- {
		root = u.middlewares[i](root)
	}

	res := root(&Context{
		Type:  UPDATE,
		Query: q,
		Model: u.model,
		Ctx:   ctx,
	})

	return res.Res.(ExecResult)
}
