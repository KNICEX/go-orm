package orm

import (
	"context"
	"database/sql"
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/KNICEX/go-orm/internal/valuer"
	"github.com/KNICEX/go-orm/model"
)

type DBOption func(db *DB)

type DB struct {
	core

	db *sql.DB
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
		core: core{
			creator: valuer.NewUnsafeValue,
			r:       model.NewRegistry(),
			dialect: &standardSQL{},
		},

		db: db,
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

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{
		tx: tx,
	}, nil
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

func (db *DB) getCore() core {
	return core{
		r:       db.r,
		dialect: db.dialect,
		creator: db.creator,
	}
}

func (db *DB) DoTx(ctx context.Context,
	fn func(ctx context.Context, tx *Tx) error,
	opts *sql.TxOptions) error {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	panicked := true
	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			err = errs.NewErrFailedToRollback(err, e)
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx, tx)
	panicked = false
	return err
}
