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
			if err := b.buildColumn(Column{name: c.arg, table: c.table}); err != nil {
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
		// 条件表达式不允许列别名
		if err := b.buildColumn(Column{name: exp.arg, table: exp.table}); err != nil {
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
		if err := b.buildJoin(t); err != nil {
			return err
		}
	case SubQuery:
		if err := b.buildSubQuery(t); err != nil {
			return err
		}
	default:
		return errs.NewErrUnsupportedTable(table)
	}
	return nil
}

func (b *builder) buildJoin(j Join) error {
	b.sb.WriteByte('(')
	// left
	if err := b.buildTable(j.left); err != nil {
		return err
	}
	b.sb.WriteByte(' ')
	b.sb.WriteString(j.typ)
	b.sb.WriteByte(' ')
	// right
	if err := b.buildTable(j.right); err != nil {
		return err
	}
	// using
	if len(j.using) > 0 {
		b.sb.WriteString(" USING (")
		for _, col := range j.using {
			if err := b.buildColumn(Column{name: col}); err != nil {
				return err
			}
			if col != j.using[len(j.using)-1] {
				b.sb.WriteByte(',')
			}
		}
		b.sb.WriteByte(')')
	}
	// on
	if len(j.on) > 0 {
		b.sb.WriteString(" ON ")
		if err := b.buildPredicate(j.on); err != nil {
			return err
		}
	}
	b.sb.WriteByte(')')
	return nil
}

func (b *builder) buildSubQuery(s SubQuery) error {
	b.sb.WriteByte('(')
	q, err := s.s.Build()
	if err != nil {
		return err
	}
	b.sb.WriteString(q.SQL[:len(q.SQL)-1]) // 去掉分号
	b.sb.WriteByte(')')
	if len(q.Args) > 0 {
		b.addArgs(q.Args...)
	}
	if s.alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(s.alias)
	}
	return nil
}

func (b *builder) buildColumn(c Column) error {
	var alias string
	table := c.table
	if table != nil {
		alias = table.tableAlias()
	}
	if alias != "" {
		b.quote(alias)
		b.sb.WriteByte('.')
	}
	colName, err := b.colName(table, c.name)
	if err != nil {
		return err
	}
	b.quote(colName)

	if c.alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(c.alias)
	}
	return nil
}

// colName 获取列名
func (b *builder) colName(table TableReference, fd string) (string, error) {
	switch tab := table.(type) {
	case nil:
		fdMeta, ok := b.model.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return fdMeta.ColName, nil
	case Table:
		m, err := b.r.Get(tab.entity)
		if err != nil {
			return "", err
		}
		fdMeta, ok := m.FieldMap[fd]
		if !ok {
			return "", errs.NewErrUnknownField(fd)
		}
		return fdMeta.ColName, nil
	case Join:
		// 先找左表，再找右表
		colName, err := b.colName(tab.left, fd)
		if err == nil {
			return colName, nil
		}
		return b.colName(tab.right, fd)
	case SubQuery:
		if len(tab.cols) > 0 {
			for _, col := range tab.cols {
				if col.selectedAlias() == fd {
					return fd, nil
				}

				if col.fieldName() == fd {
					return b.colName(col.target(), fd)
				}
			}
			return "", errs.NewErrUnknownField(fd)
		}
		return b.colName(tab.table, fd)
	default:
		return "", errs.NewErrUnsupportedTable(table)
	}
}
