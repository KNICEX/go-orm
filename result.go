package orm

import "database/sql"

type ExecResult struct {
	err error
	res sql.Result
}

func (r ExecResult) LastInsertId() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.LastInsertId()
}

func (r ExecResult) RowsAffected() (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.res.RowsAffected()
}

func (r ExecResult) Err() error {
	return r.err
}
