package orm

import (
	"context"
	"database/sql"
	"math/rand"
)

type ShardingDB struct {
	Shards map[string]*MasterSlaveDB
}

type MasterSlaveDB struct {
	core
	Master *sql.DB
	Slaves []*sql.DB
}

func (m *MasterSlaveDB) getCore() *core {
	return m.getCore()
}

func (m *MasterSlaveDB) pick() int {
	return rand.Intn(len(m.Slaves))
}

func (m *MasterSlaveDB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return m.Slaves[m.pick()].QueryContext(ctx, query, args...)
}

func (m *MasterSlaveDB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return m.Master.ExecContext(ctx, query, args...)
}

func (m *MasterSlaveDB) queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return m.Slaves[m.pick()].QueryRowContext(ctx, query, args...)
}
