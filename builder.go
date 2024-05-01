package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/KNICEX/go-orm/model"
	"strings"
)

type builder struct {
	args  []any
	sb    strings.Builder
	model *model.Model

	dialect Dialect
	quoter  byte
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) addArgs(vals ...any) {
	if len(vals) == 0 {
		return
	}
	if b.args == nil {
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, vals...)
}

// buildColumns 构造查询列
func (b *builder) buildColumns(cols []Selectable) error {
	if len(cols) == 0 {
		b.sb.WriteByte('*')
		return nil
	}
	for i, col := range cols {
		switch c := col.(type) {
		case Column:
			if err := b.buildColumn(c); err != nil {
				return err
			}
		case Aggregate:
			b.sb.WriteString(c.fn)

			b.sb.WriteByte('(')
			if err := b.buildColumn(Column{name: c.arg}); err != nil {
				return err
			}
			b.sb.WriteByte(')')

			if c.alias != "" {
				b.sb.WriteString(" AS ")
				b.quote(c.alias)
			}
		case RawExpr:
			b.sb.WriteString(c.raw)
			b.addArgs(c.args...)
		}

		if i != len(cols)-1 {
			b.sb.WriteByte(',')
		}
	}

	return nil
}

// buildPredicate 构造谓词（条件）
func (b *builder) buildPredicate(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return b.buildExpression(p)
}

// buildExpression 递归构造条件表达式
func (b *builder) buildExpression(expr Expression) error {
	switch exp := expr.(type) {
	case nil:
		return nil
	case Predicate:

		// 如果左边也是一个表达式，那么需要加括号
		_, ok := exp.left.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}

		// 操作符
		if exp.op != "" {
			b.sb.WriteByte(' ')
			b.sb.WriteString(exp.op.String())
			b.sb.WriteByte(' ')
		}

		// 如果右边也是一个表达式，那么需要加括号
		_, ok = exp.right.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}
	case Column:
		// 条件表达式不允许列别名
		exp.alias = ""
		return b.buildColumn(Column{name: exp.name})
	case value:
		b.sb.WriteByte('?')
		b.addArgs(exp.val)
	case RawExpr:
		b.sb.WriteByte('(')
		b.sb.WriteString(exp.raw)
		b.sb.WriteByte(')')
		b.addArgs(exp.args...)
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

// buildColumn 构造单个列
func (b *builder) buildColumn(c Column) error {
	fd, ok := b.model.FieldMap[c.name]
	if !ok {
		return errs.NewErrUnknownField(c.name)
	}

	b.quote(fd.ColName)

	if c.alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(c.alias)
	}
	return nil
}
