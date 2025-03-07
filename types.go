package orm

import (
	"context"
)

// Querier 用于查询
type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// Executor 用于增删改
type Executor interface {
	Exec(ctx context.Context) ExecResult
}

type SqlBuilder interface {
	Build() (*Query, error)
}

type Query struct {
	SQL      string
	Database string
	Args     []any
}
