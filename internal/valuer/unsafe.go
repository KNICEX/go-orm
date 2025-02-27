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

	// 仅用于Field方法
	addr unsafe.Pointer
}

var _ Creator = NewUnsafeValue

// NewUnsafeValue
// val 是结构体一级指针或者切片一级指针
func NewUnsafeValue(model *model.Model, val any) Value {
	return &unsafeValue{
		model: model,
		val:   val,
		addr:  reflect.ValueOf(val).UnsafePointer(),
	}
}

func (u *unsafeValue) Field(name string) (any, error) {
	fd, ok := u.model.FieldMap[name]
	if !ok {
		return nil, errs.NewErrUnknownColumn(name)
	}
	fdPtr := unsafe.Pointer(uintptr(u.addr) + fd.Offset)
	return reflect.NewAt(fd.Typ, fdPtr).Elem().Interface(), nil
}

func (u *unsafeValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	tpValue := reflect.ValueOf(u.val).Elem()
	if tpValue.Kind() == reflect.Slice {
		// 切片元素一定是结构体一级指针
		entityType := tpValue.Type().Elem()
		entity := reflect.New(entityType.Elem()).Interface()
		// 调用该方法前会调用rows.Next()，所以这里第一行不需要再调用
		if err = u.setRow(cs, rows, entity); err != nil {
			return err
		}
		tpValue.Set(reflect.Append(tpValue, reflect.ValueOf(entity)))
		for rows.Next() {
			entity = reflect.New(entityType.Elem()).Interface()
			if err = u.setRow(cs, rows, entity); err != nil {
				return err
			}
			tpValue.Set(reflect.Append(tpValue, reflect.ValueOf(entity)))
		}
		return nil
	} else {
		return u.setRow(cs, rows, u.val)
	}
}

func (u *unsafeValue) setRow(row []string, scanner *sql.Rows, entity any) error {
	vals := make([]any, 0, len(row))
	addr := reflect.ValueOf(entity).UnsafePointer()
	for _, c := range row {
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
	return scanner.Scan(vals...)
}
