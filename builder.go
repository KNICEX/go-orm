package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
	model2 "github.com/KNICEX/go-orm/model"
	"strings"
)

type builder struct {
	args  []any
	sb    strings.Builder
	model *model2.Model
}

func (b *builder) buildPredicate(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return b.buildExpression(p)
}

// buildExpression 递归构造表达式
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
		b.sb.WriteByte(' ')
		b.sb.WriteString(exp.op.String())
		b.sb.WriteByte(' ')

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
		b.sb.WriteByte('`')
		colName, ok := b.model.FieldMap[exp.name]
		if !ok {
			return errs.NewErrUnknownField(exp.name)
		}
		b.sb.WriteString(colName.ColName)
		b.sb.WriteByte('`')
	case value:
		b.sb.WriteByte('?')
		b.args = append(b.args, exp.val)
	default:
		return errs.NewErrUnsupportedExpression(expr)
	}
	return nil
}
