package orm

type op string

const (
	opEq  op = "="
	opNot op = "NOT"
	opAnd op = "AND"
	opOr  op = "OR"
	opGt  op = ">"
	opLt  op = "<"
	opGe  op = ">="
	opLe  op = "<="
)

func (o op) String() string {
	return string(o)
}

// Expression 标记接口，用于表示一个表达式
type Expression interface {
	expr()
}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (p Predicate) expr() {}

type Column struct {
	name string
}

func (c Column) expr() {}

func Col(name string) Column {
	return Column{name: name}
}

func (c Column) Eq(arg any) Predicate {
	return Predicate{
		left: c,
		op:   opEq,
		right: value{
			val: arg,
		},
	}
}

func (c Column) Gt(arg any) Predicate {
	return Predicate{
		left: c,
		op:   opGt,
		right: value{
			val: arg,
		},
	}
}

func (c Column) Lt(arg any) Predicate {
	return Predicate{
		left: c,
		op:   opLt,
		right: value{
			val: arg,
		},
	}
}

func (c Column) Ge(arg any) Predicate {
	return Predicate{
		left: c,
		op:   opGe,
		right: value{
			val: arg,
		},
	}
}

func (c Column) Le(arg any) Predicate {
	return Predicate{
		left: c,
		op:   opLe,
		right: value{
			val: arg,
		},
	}
}

func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

// And Col("id").Eq(1).And(Col("name").Eq("foo"))
func (p Predicate) And(p2 Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opAnd,
		right: p2,
	}
}

func (p Predicate) Or(p2 Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    opOr,
		right: p2,
	}
}

type value struct {
	val any
}

func (v value) expr() {}
