package orm

// Expression 标记接口，用于表示一个表达式
type Expression interface {
	expr()
}

// RawExpr 表示一个原始表达式
type RawExpr struct {
	raw  string
	args []any
}

func (r RawExpr) selectedAlias() string { return "" }

func (r RawExpr) target() TableReference {
	return nil
}

func (r RawExpr) fieldName() string {
	return ""
}
func (r RawExpr) expr()      {}
func (r RawExpr) orderAble() {}
func (r RawExpr) groupAble() {}
func (r RawExpr) setAble()   {}

// Raw 创建一个原始表达式
// Raw 不会对表达式进行任何处理，直接拼接到 SQL 语句中
// 谨慎使用
func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}
