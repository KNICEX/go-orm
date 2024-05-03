package orm

import (
	"github.com/KNICEX/go-orm/internal/valuer"
	"github.com/KNICEX/go-orm/model"
)

type core struct {
	model   *model.Model
	dialect Dialect
	creator valuer.Creator
	r       model.Registry

	middlewares []Middleware
}
