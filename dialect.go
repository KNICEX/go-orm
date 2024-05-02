package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"strconv"
)

var (
	DialectMySQL    = &mysqlDialect{}
	DialectPostgres = &postgresDialect{}
	DialectSQLite3  = &sqlite3Dialect{}
)

var dialectMap = map[string]Dialect{
	"mysql":    DialectMySQL,
	"postgres": DialectPostgres,
	"sqlite3":  DialectSQLite3,
}

type Dialect interface {
	// quoter 字段引号
	quoter() byte
	buildUpsert(sb *builder, upsert *Upsert) error
	buildOffsetLimit(sb *builder, offset, limit int) error
}

type standardSQL struct {
}

func (s *standardSQL) quoter() byte {
	return '`'
}

func (s *standardSQL) buildUpsert(b *builder, odk *Upsert) error {
	panic("not implemented")
}

func (s *standardSQL) buildOffsetLimit(b *builder, offset, limit int) error {
	if limit > 0 {
		b.sb.WriteString(" LIMIT ")
		b.sb.WriteString(strconv.Itoa(limit))
	}
	if offset > 0 {
		b.sb.WriteString(" OFFSET ")
		b.sb.WriteString(strconv.Itoa(offset))
	}
	return nil
}

type mysqlDialect struct {
	standardSQL
}

func (s *mysqlDialect) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for i, assign := range upsert.assigns {
		if i > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString(" = ?")
			b.addArgs(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString(" = VALUES(")
			b.quote(fd.ColName)
			b.sb.WriteByte(')')
		default:
			return errs.NewErrUnsupportedAssignable(a)
		}

	}
	return nil
}

type sqlite3Dialect struct {
	standardSQL
}

func (s *sqlite3Dialect) buildUpsert(b *builder, upsert *Upsert) error {
	b.sb.WriteString(" ON CONFLICT(")
	for i, col := range upsert.conflictColumns {
		if i > 0 {
			b.sb.WriteByte(',')
		}
		if err := b.buildColumn(Column{name: col}); err != nil {
			return err
		}
	}
	b.sb.WriteString(") DO UPDATE SET ")

	for i, assign := range upsert.assigns {
		if i > 0 {
			b.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			fd, ok := b.model.FieldMap[a.col]
			if !ok {
				return errs.NewErrUnknownField(a.col)
			}
			b.quote(fd.ColName)
			b.sb.WriteString(" = ?")
			b.addArgs(a.val)
		case Column:
			fd, ok := b.model.FieldMap[a.name]
			if !ok {
				return errs.NewErrUnknownField(a.name)
			}
			b.quote(fd.ColName)
			b.sb.WriteString(" = EXCLUDED.")
			b.quote(fd.ColName)
		default:
			return errs.NewErrUnsupportedAssignable(a)
		}

	}
	return nil
}

type postgresDialect struct {
	standardSQL
}
