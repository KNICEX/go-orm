package orm

import "context"

type RawQuerier[T any] struct {
	sess Session
	*core
	sql  string
	args []any
}

func RawQuery[T any](sess Session, query string, args ...any) *RawQuerier[T] {
	return &RawQuerier[T]{
		sql:  query,
		args: args,
		sess: sess,
		core: sess.getCore(),
	}
}

func (r *RawQuerier[T]) Build() (*Query, error) {

	return &Query{
		SQL:  r.sql,
		Args: r.args,
	}, nil
}

func (r *RawQuerier[T]) Exec(ctx context.Context) ExecResult {
	return exec(ctx, r, r.sess, r.core, RAW)
}

func (r *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	// 获取模型
	m, err := r.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	r.model = m

	resEntity := new(T)
	err = get(ctx, r, r.sess, r.core, RAW, resEntity)
	if err != nil {
		return nil, err

	}
	return resEntity, nil
}

func (r *RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	// 获取模型
	m, err := r.r.Get(new(T))
	if err != nil {
		return nil, err
	}
	r.model = m

	resEntity := new([]*T)
	err = get(ctx, r, r.sess, r.core, RAW, resEntity)
	if err != nil {
		return nil, err
	}
	return *resEntity, nil
}
