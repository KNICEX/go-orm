package orm

type TableReference interface {
	table()
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

func (t Table) table() {
	//TODO implement me
	panic("implement me")
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

func (j Join) table() {
	//TODO implement me
	panic("implement me")
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
