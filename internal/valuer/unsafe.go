package valuer

import (
	"database/sql"
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/KNICEX/go-orm/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	model *model.Model
	val   any
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(model *model.Model, val any) Value {
	return &unsafeValue{
		model: model,
		val:   val,
	}
}

func (u *unsafeValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	vals := make([]any, 0, len(cs))

	// 起始地址
	addr := reflect.ValueOf(u.val).UnsafePointer()
	for _, c := range cs {
		fd, ok := u.model.ColMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// 字段地址
		fdAddr := unsafe.Pointer(uintptr(addr) + fd.Offset)
		// 指针类型
		val := reflect.NewAt(fd.Typ, fdAddr).Interface()
		vals = append(vals, val)
	}
	return rows.Scan(vals...)
}
