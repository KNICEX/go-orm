package orm

import (
	"database/sql"
	"github.com/KNICEX/go-orm/internal/valuer"
	"github.com/KNICEX/go-orm/model"
)

type DBOption func(db *DB)

type DB struct {
	r       model.Registry
	db      *sql.DB
	creator valuer.Creator
	dialect Dialect
}

func Open(driver, dsn string, ops ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	res, err := OpenDB(db, ops...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func OpenDB(db *sql.DB, opts ...DBOption) (*DB, error) {
	res := &DB{
		r:       model.NewRegistry(),
		db:      db,
		creator: valuer.NewUnsafeValue,
		dialect: &standardSQL{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = valuer.NewReflectValue
	}
}

func DBWithRegistry(r model.Registry) DBOption {
	return func(db *DB) {
		db.r = r
	}
}

func DBWithDialect(d Dialect) DBOption {
	return func(db *DB) {
		db.dialect = d
	}
}

func MustOpen(driver, dsn string, ops ...DBOption) *DB {
	db, err := Open(driver, dsn, ops...)
	if err != nil {
		panic(err)
	}
	return db
}
