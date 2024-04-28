package valuer

import (
	"database/sql"
	"github.com/KNICEX/go-orm/model"
)

type Value interface {
	SetColumns(rows *sql.Rows) error
}

type Creator func(model *model.Model, entity any) Value
