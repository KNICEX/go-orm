package valuer

import (
	"database/sql"
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/KNICEX/go-orm/model"
	"reflect"
)

type reflectValue struct {
	model *model.Model
	val   reflect.Value
}

var _ Creator = NewReflectValue

// NewReflectValue
// val 结构体一级指针或slice一级指针
func NewReflectValue(model *model.Model, val any) Value {
	return &reflectValue{
		model: model,
		val:   reflect.ValueOf(val).Elem(),
	}
}

func (r *reflectValue) Field(name string) (any, error) {
	return r.val.FieldByName(name).Interface(), nil
}

func (r *reflectValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	if r.val.Kind() == reflect.Slice {
		// 切片元素一定是结构体一级指针
		entityType := r.val.Type().Elem()
		entity := reflect.New(entityType.Elem()).Interface()
		// 调用该方法前会调用rows.Next()，所以这里第一行不需要再调用
		if err = r.setRow(cs, rows, entity); err != nil {
			return err
		}
		r.val.Set(reflect.Append(r.val, reflect.ValueOf(entity)))
		for rows.Next() {
			entity = reflect.New(entityType.Elem()).Interface()
			if err = r.setRow(cs, rows, entity); err != nil {
				return err
			}
			r.val.Set(reflect.Append(r.val, reflect.ValueOf(entity)))
		}
		return nil
	} else {
		return r.setRow(cs, rows, r.val.Addr().Interface())
	}
}

// row 数据库行字段
// entity 结构体一级指针
func (r *reflectValue) setRow(row []string, scanner *sql.Rows, entity any) error {
	vals := make([]any, 0, len(row))
	valElems := make([]reflect.Value, 0, len(row))
	for _, c := range row {
		fd, ok := r.model.ColMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		// 放射创建字段对应类型的指针， 用于 Scan
		val := reflect.New(fd.Typ)
		vals = append(vals, val.Interface())
		valElems = append(valElems, val.Elem())
	}

	err := scanner.Scan(vals...)
	if err != nil {
		return err
	}

	tpValue := reflect.ValueOf(entity).Elem()
	for i, c := range row {
		fd, _ := r.model.ColMap[c]
		tpValue.FieldByName(fd.GoName).Set(valElems[i])
	}
	return nil
}
