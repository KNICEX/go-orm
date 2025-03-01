package orm

const (
	asc  = "ASC"
	desc = "DESC"
)

type Column struct {
	table TableReference

	name  string
	alias string
	desc  bool
}

// 实现 Expression 标记接口， Column 可以作为一个表达式
func (c Column) expr() {}

// 实现 selectable 标记接口， Column 可以作为一个SELECT语句的列
func (c Column) selectedAlias() string { return c.alias }
func (c Column) target() TableReference {
	return c.table
}

func (c Column) fieldName() string {
	return c.name
}

// 实现 assignable 标记接口， Column 可以作为一个ON DUPLICATE KEY UPDATE语句的列
func (c Column) assign() {}

// 实现 orderBy 标记接口， Column 可以作为一个ORDER BY语句的列
func (c Column) orderAble() {}

// 实现 groupBy 标记接口， Column 可以作为一个GROUP BY语句的列
func (c Column) groupAble() {}

func Col(name string) Column {
	return Column{name: name}
}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
		table: c.table,
	}
}

func valueOf(arg any) Expression {
	switch val := arg.(type) {
	case Expression:
		return val
	default:
		return value{val: arg}
	}
}

// Eq
// 用法： Col("id").Eq(1)
func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: valueOf(arg),
	}
}

func (c Column) Gt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGt,
		right: valueOf(arg),
	}
}

func (c Column) Lt(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLt,
		right: valueOf(arg),
	}
}

func (c Column) Ge(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opGe,
		right: valueOf(arg),
	}
}

func (c Column) Le(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLe,
		right: valueOf(arg),
	}
}

func (c Column) Like(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    opLike,
		right: valueOf(arg),
	}
}

func (c Column) In(args ...any) Predicate {
	return Predicate{
		left:  c,
		op:    opIn,
		right: valueOf(args),
	}
}

func (c Column) NotIn(args ...any) Predicate {
	return Predicate{
		left:  c,
		op:    opNotIn,
		right: valueOf(args),
	}
}

func (c Column) Asc() Column {
	return Column{
		name:  c.name,
		table: c.table,
	}
}

func (c Column) Desc() Column {
	return Column{
		name:  c.name,
		desc:  true,
		table: c.table,
	}
}
