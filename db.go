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
	}
	for _, op := range opts {
		op(res)
	}
	return res, nil
}

func DBUseReflect() DBOption {
	return func(db *DB) {
		db.creator = valuer.NewReflectValue
	}
}

func MustOpen(driver, dsn string, ops ...DBOption) *DB {
	db, err := Open(driver, dsn, ops...)
	if err != nil {
		panic(err)
	}
	return db
}
