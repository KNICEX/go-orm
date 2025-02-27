package orm

type TableReference interface {
	tableAlias() string
}

type Table struct {
	entity any
	alias  string
}

type joinType = string

const (
	innerJoin joinType = "INNER JOIN"
	leftJoin  joinType = "LEFT JOIN"
	rightJoin joinType = "RIGHT JOIN"
)

func (t Table) tableAlias() string {
	return t.alias
}

func TableOf(entity any) Table {
	return Table{
		entity: entity,
	}
}

func (t Table) Col(name string) Column {
	return Column{
		table: t,
		name:  name,
	}
}

func (t Table) As(alias string) Table {
	t.alias = alias
	return t
}

func (t Table) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   innerJoin,
	}
}

func (t Table) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   leftJoin,
	}
}

func (t Table) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: right,
		typ:   rightJoin,
	}
}

type Join struct {
	left  TableReference
	right TableReference
	typ   joinType

	on    []Predicate
	using []string
}

func (j Join) tableAlias() string {
	return ""
}

func (j Join) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   innerJoin,
	}
}

func (j Join) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   leftJoin,
	}
}

func (j Join) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: right,
		typ:   rightJoin,
	}
}

type JoinBuilder struct {
	left  TableReference
	right TableReference
	typ   joinType
}

func (j *JoinBuilder) On(ps ...Predicate) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		on:    ps,
	}
}

func (j *JoinBuilder) Using(cols ...string) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		using: cols,
	}
}

type SubQuery struct {
	table TableReference
	s     SqlBuilder
	cols  []Selectable
	alias string
}

func (s SubQuery) tableAlias() string {
	return s.alias
}

func (s SubQuery) Col(name string) Column {
	return Column{
		table: s,
		name:  name,
	}
}

func (s SubQuery) Max(col string) Aggregate {
	return Aggregate{
		table: s,
		fn:    "MAX",
		arg:   col,
	}
}

func (s SubQuery) Min(col string) Aggregate {
	return Aggregate{
		table: s,
		fn:    "MIN",
		arg:   col,
	}
}

func (s SubQuery) Count(col string) Aggregate {
	return Aggregate{
		table: s,
		fn:    "COUNT",
		arg:   col,
	}
}

func (s SubQuery) Sum(col string) Aggregate {
	return Aggregate{
		table: s,
		fn:    "SUM",
		arg:   col,
	}
}

func (s SubQuery) Avg(col string) Aggregate {
	return Aggregate{
		table: s,
		fn:    "AVG",
		arg:   col,
	}
}

func (s SubQuery) Join(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: right,
		typ:   innerJoin,
	}
}

func (s SubQuery) LeftJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: right,
		typ:   leftJoin,
	}
}

func (s SubQuery) RightJoin(right TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: right,
		typ:   rightJoin,
	}
}
