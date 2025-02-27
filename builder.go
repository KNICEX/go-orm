package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"strings"
)

type builder struct {
	*core
	args []any
	sb   strings.Builder

	quoter byte
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
		return b.buildColumn(exp)
	case value:
		b.sb.WriteByte('?')
		b.addArgs(exp.val)
	case RawExpr:
		b.sb.WriteByte('(')
		b.sb.WriteString(exp.raw)
		b.sb.WriteByte(')')
		b.addArgs(exp.args...)
	case Aggregate:
		b.sb.WriteString(exp.fn)
		b.sb.WriteByte('(')
		if err := b.buildColumn(Column{name: exp.arg}); err != nil {
			return err
		}
		b.sb.WriteByte(')')
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}

func (b *builder) buildTable(table TableReference) error {
	switch t := table.(type) {
	case nil:
		// 没有调用From
		b.quote(b.model.TableName)
	case Table:
		m, err := b.r.Get(t.entity)
		if err != nil {
			return err
		}
		b.quote(m.TableName)
		if t.alias != "" {
			b.sb.WriteString(" AS ")
			b.quote(t.alias)
		}
	case Join:
		b.sb.WriteByte('(')
		// left
		if err := b.buildTable(t.left); err != nil {
			return err
		}
		b.sb.WriteByte(' ')
		b.sb.WriteString(t.typ)
		b.sb.WriteByte(' ')
		// right
		if err := b.buildTable(t.right); err != nil {
			return err
		}
		// using
		if len(t.using) > 0 {
			b.sb.WriteString(" USING (")
			for _, col := range t.using {
				if err := b.buildColumn(Column{name: col}); err != nil {
					return err
				}
				if col != t.using[len(t.using)-1] {
					b.sb.WriteByte(',')
				}
			}
			b.sb.WriteByte(')')
		}
		// on
		if len(t.on) > 0 {
			b.sb.WriteString(" ON ")
			if err := b.buildPredicate(t.on); err != nil {
				return err
			}
		}
		b.sb.WriteByte(')')
	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
}

// buildColumn 构造单个列
func (b *builder) buildColumn(c Column) error {
	switch table := c.table.(type) {
	case nil:
		fd, ok := b.model.FieldMap[c.name]
		if !ok {
			return errs.NewErrUnknownField(c.name)
		}

		b.quote(fd.ColName)

	case Table:
		m, err := b.r.Get(table.entity)
		if err != nil {
			return err
		}
		fd, ok := m.FieldMap[c.name]
		if !ok {
			return errs.NewErrUnknownField(c.name)
		}
		if table.alias != "" {
			b.quote(table.alias)
		} else {
			b.quote(m.TableName)
		}
		b.sb.WriteByte('.')
		b.quote(fd.ColName)

	default:
		return errs.NewErrUnsupportedTable(c.table)
	}
	if c.alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(c.alias)
	}
	return nil

}
