package valuer

import (
	"database/sql"
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/KNICEX/go-orm/model"
	"reflect"
)

type reflectValue struct {
	model *model.Model
	val   any
}

var _ Creator = NewReflectValue

func NewReflectValue(model *model.Model, val any) Value {
	return &reflectValue{
		model: model,
		val:   val,
	}
}

func (r *reflectValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		fd, ok := r.model.ColMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// 放射创建字段对应类型的指针， 用于 Scan
		val := reflect.New(fd.Typ)
		vals = append(vals, val.Interface())
		valElems = append(valElems, val.Elem())
	}

	err = rows.Scan(vals...)
	if err != nil {
		return err
	}
	tpValue := reflect.ValueOf(r.val).Elem()
	for i, c := range cs {
		// 这里不用检查是否存在，因为上面已经检查过了
		fd, _ := r.model.ColMap[c]
		tpValue.FieldByName(fd.GoName).Set(valElems[i])
	}
	return nil
}
