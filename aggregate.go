package orm

// Aggregate 聚合函数
// AVG, COUNT, MAX, MIN, SUM
type Aggregate struct {
	table TableReference
	fn    string
	arg   string
	alias string
}

func (a Aggregate) expr() {}

func (a Aggregate) selectedAlias() string { return a.alias }

func (a Aggregate) fieldName() string {
	return a.arg
}

func (a Aggregate) target() TableReference {
	return a.table
}

func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		table: a.table,
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

func Avg(col string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: col,
	}
}

func Count(col string) Aggregate {
	return Aggregate{
		fn:  "COUNT",
		arg: col,
	}
}

func Max(col string) Aggregate {
	return Aggregate{
		fn:  "MAX",
		arg: col,
	}
}

func Min(col string) Aggregate {
	return Aggregate{
		fn:  "MIN",
		arg: col,
	}
}

func Sum(col string) Aggregate {
	return Aggregate{
		fn:  "SUM",
		arg: col,
	}
}

func (a Aggregate) Gt(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opGt,
		right: valueOf(arg),
	}
}

func (a Aggregate) Lt(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opLt,
		right: valueOf(arg),
	}
}

func (a Aggregate) Ge(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opGe,
		right: valueOf(arg),
	}
}

func (a Aggregate) Le(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opLe,
		right: valueOf(arg),
	}
}

func (a Aggregate) Eq(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opEq,
		right: valueOf(arg),
	}
}
