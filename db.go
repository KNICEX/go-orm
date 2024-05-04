package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/KNICEX/go-orm/internal/valuer"
	"github.com/KNICEX/go-orm/model"
	"log"
	"time"
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

func DBWithMiddlewares(middlewares ...Middleware) DBOption {
	return func(db *DB) {
		db.middlewares = append(db.middlewares, middlewares...)
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

func (db *DB) queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.db.QueryRowContext(ctx, query, args...)
}

func (db *DB) getCore() *core {
	return &core{
		dialect:     db.dialect,
		creator:     db.creator,
		r:           db.r,
		middlewares: db.middlewares,
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

func (db *DB) Wait() error {
	err := db.db.Ping()
	for errors.Is(err, driver.ErrBadConn) {
		log.Println("waiting for database start...")
		err = db.db.Ping()
		time.Sleep(time.Second * 3)
	}
	return err
}

// HasTable 判断表是否存在
func (db *DB) HasTable(tableName string) bool {
	existSql, args := db.dialect.TableExistSQL(tableName)
	res := db.queryRowContext(context.Background(), existSql, args...)
	if res.Err() != nil {
		return false
	}
	var tempName string
	err := res.Scan(&tempName)
	if err != nil {
		return false
	}
	return tempName == tableName
}

// CreateTable 只接受结构体一级指针作为参数
func (db *DB) CreateTable(values ...any) error {
	for _, val := range values {
		m, err := db.r.Get(val)
		if err != nil {
			return err
		}
		if db.HasTable(m.TableName) {
			return errs.NewErrTableExist(m.TableName)
		}
		//	TODO 创建表
	}
	return nil
}

// Migrate 只接受结构体一级指针作为参数
// 只修改表结构，不会创建表
func (db *DB) Migrate(values ...any) error {
	return db.DoTx(context.Background(), func(ctx context.Context, tx *Tx) error {
		for _, val := range values {
			m, err := db.r.Get(val)
			if err != nil {
				return err
			}
			if !tx.HasTable(m.TableName) {
				return errs.NewErrTableNotExist(m.TableName)
			}

			//	TODO 修改表结构

		}
		return nil
	}, nil)
}

// AutoMigrate 只接受结构体一级指针作为参数
// 如果表不存在则创建表，如果表存在则修改表结构
func (db *DB) AutoMigrate(values ...any) error {
	return db.DoTx(context.Background(), func(ctx context.Context, tx *Tx) error {
		for _, val := range values {
			m, err := db.r.Get(val)
			if err != nil {
				return err
			}
			if !tx.HasTable(m.TableName) {
				err = tx.CreateTable(val)
				if err != nil {
					return err
				}
			}
			err = tx.Migrate(val)
			if err != nil {
				return err
			}
		}
		return nil
	}, nil)
}
