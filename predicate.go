package orm

type op string

const (
	opEq    op = "="
	opNot   op = "NOT"
	opAnd   op = "AND"
	opOr    op = "OR"
	opGt    op = ">"
	opLt    op = "<"
	opGe    op = ">="
	opLe    op = "<="
	opLike  op = "LIKE"
	opIn    op = "IN"
	opNotIn op = "NOT IN"
)

func (o op) String() string {
	return string(o)
}

// Predicate 谓词，用于构造WHERE条件, 例如 Col("id").Eq(1)
type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (p Predicate) expr() {}

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
