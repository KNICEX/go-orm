package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
)

var (
	DialectMySQL    = &mysqlDialect{}
	DialectPostgres = &postgresDialect{}
	DialectSQLite   = &sqliteDialect{}
)

var dialectMap = map[string]Dialect{
	"mysql":    DialectMySQL,
	"postgres": DialectPostgres,
	"sqlite":   DialectSQLite,
}

type Dialect interface {
	// quoter 字段引号
	quoter() byte
	buildOnDuplicateKey(sb *builder, odk *OnDuplicateKey) error
}

type standardSQL struct {
}

func (s *standardSQL) quoter() byte {
	return '`'
}

func (s *standardSQL) buildOnDuplicateKey(b *builder, odk *OnDuplicateKey) error {
	b.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for i, assign := range odk.assigns {
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

type mysqlDialect struct {
	standardSQL
}

type postgresDialect struct {
	standardSQL
}

type sqliteDialect struct {
	standardSQL
}
