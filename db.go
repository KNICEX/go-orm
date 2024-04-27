package orm

import "reflect"

type DBOption func(db *DB)

type DB struct {
	r *registry
}

func NewDB(ops ...DBOption) (*DB, error) {
	res := &DB{
		r: &registry{
			models: make(map[reflect.Type]*model, 24),
		},
	}
	for _, op := range ops {
		op(res)
	}
	return res, nil
}
