package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KNICEX/go-orm/internal/errs"
	"golang.org/x/sync/errgroup"
)

type ShardingSelector[T any] struct {
	table    TableReference
	where    []Predicate
	columns  []Selectable
	orderBys []OrderAble
	groupBys []GroupAble
	having   []Predicate
	// 只查询COUNT
	count  bool
	offset int
	limit  int

	builder
	db *ShardingDB
}

type Dst struct {
	Database string
	Table    string
}

func NewShardingSelector[T any](db *ShardingDB) *ShardingSelector[T] {
	return &ShardingSelector[T]{
		db: db,
	}
}

func (s *ShardingSelector[T]) findDst() ([]Dst, error) {
	// 分析where条件，找到所有目标表
	if len(s.where) == 0 {
		return nil, errors.New("orm: no where")
	}
	p := s.where[0]
	for i := 1; i < len(s.where); i++ {
		p = p.And(s.where[i])
	}
	return s.findDstByPredicate(p)
}

func (s *ShardingSelector[T]) merge(left, right []Dst) []Dst {
	res := make([]Dst, 0, len(left)+len(right))
	for _, l := range left {
		res = append(res, l)
	}
	for _, r := range right {
		exist := false
		for _, l := range left {
			if l.Database == r.Database && l.Table == r.Table {
				exist = true
				break
			}
		}
		if !exist {
			res = append(res, r)
		}
	}
	return res
}

func (s *ShardingSelector[T]) findDstByPredicate(p Predicate) ([]Dst, error) {
	var res []Dst
	switch p.op {
	case opAnd:
		right, err := s.findDstByPredicate(p.right.(Predicate))
		if err != nil {
			return nil, err
		}
		if len(right) == 0 {
			return s.findDstByPredicate(p.left.(Predicate))
		}
		left, err := s.findDstByPredicate(p.left.(Predicate))
		if err != nil {
			return nil, err
		}
		if len(left) == 0 {
			return right, nil
		}
		return s.merge(left, right), nil
	case opNot:
	case opOr:
	case opEq:
		left, ok := p.left.(Column)
		if ok {
			_, isSk := s.model.Sks[left.name]
			right, ok := p.right.(value)
			if !ok {
				return nil, errors.New("orm: 暂不支持复杂条件")
			}
			if isSk {
				db, tbl := s.model.Sf(map[string]any{
					left.name: right.val,
				})
				res = append(res, Dst{
					Database: db,
					Table:    tbl,
				})
			}
		}
	case opGe:
	case opGt:
	case opLe:
	case opLt:
	case opIn:
	case opNotIn:
	case opLike:
	default:
		return nil, errors.New("orm: unsupported sharding select operator")
	}
	return res, nil
}
func (s *ShardingSelector[T]) Build() ([]*Query, error) {
	m, err := s.r.Register(new(T))
	if err != nil {
		return nil, err
	}
	s.model = m

	dst, err := s.findDst()
	if err != nil {
		return nil, err
	}
	// 生成SQL
	queries := make([]*Query, 0, len(dst))
	for _, d := range dst {
		q, err := s.build(d.Database, d.Table)
		if err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}
	return queries, nil
}
func (s *ShardingSelector[T]) build(database, table string) (*Query, error) {
	var err error
	s.sb.WriteString("SELECT ")

	if s.count {
		// 只查询 COUNT
		s.sb.WriteString("COUNT(*)")
	} else if err = s.buildColumns(s.columns); err != nil {
		return nil, err
	}

	s.sb.WriteString(" FROM ")
	// 表名 如果没有指定表名，则使用类型名
	s.sb.WriteString(fmt.Sprintf("%s.%s", database, table))

	// 条件构造
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		if err := s.buildPredicate(s.where); err != nil {
			return nil, err
		}
	}

	// 排序
	if len(s.orderBys) > 0 {
		s.sb.WriteString(" ORDER BY ")
		for i, ob := range s.orderBys {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			switch o := ob.(type) {
			case Column:
				fd, ok := s.model.FieldMap[o.name]
				if !ok {
					return nil, errs.NewErrUnknownField(o.name)
				}
				s.quote(fd.ColName)
				if o.desc {
					s.sb.WriteString(" DESC")
				} else {
					s.sb.WriteString(" ASC")
				}
			case RawExpr:
				s.sb.WriteByte('(')
				s.sb.WriteString(o.raw)
				s.sb.WriteByte(')')
			}
		}
	}

	// 分组
	if len(s.groupBys) > 0 {
		s.sb.WriteString(" GROUP BY ")
		for i, gb := range s.groupBys {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			switch g := gb.(type) {
			case Column:
				if err = s.buildColumn(g); err != nil {
					return nil, err
				}
			case RawExpr:
				s.sb.WriteString(g.raw)
			}
		}

		if len(s.having) > 0 {
			// having
			s.sb.WriteString(" HAVING ")
			if err = s.buildPredicate(s.having); err != nil {
				return nil, err
			}
		}

	}

	// limit offset
	if err = s.dialect.buildOffsetLimit(&s.builder, s.offset, s.limit); err != nil {
		return nil, err
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:      s.sb.String(),
		Args:     s.args,
		Database: database,
	}, nil
}

func (s *ShardingSelector[T]) Get(ctx context.Context) (T, error) {
	var res T
	qs, err := s.Build()
	if err != nil {
		return res, err
	}
	eg := errgroup.Group{}
	rowSlice := make([]*sql.Rows, len(qs))
	for i, q := range qs {
		eg.Go(func() error {
			db, ok := s.db.Shards[q.Database]
			if !ok {
				return errors.New("orm: unknown database")
			}
			row, err := db.queryContext(ctx, q.SQL, q.Args...)
			if err != nil {
				return err
			}
			rowSlice[i] = row
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return res, err
	}
	for _, row := range rowSlice {
		if row.Next() {
			// 本身只要返回一条记录，所以就用最后一个有值的结果
			err = s.core.creator(s.core.model, &res).SetColumns(row)
			if err != nil {
				return res, err
			}
		}
	}
	return res, nil
}
