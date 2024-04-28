package unsafe

import (
	"errors"
	"reflect"
	"unsafe"
)

type Accessor struct {
	fields  map[string]*FieldMeta
	address unsafe.Pointer
}

type FieldMeta struct {
	Offset uintptr
	typ    reflect.Type
}

// NewAccessor 只支持 struct 指针
func NewAccessor(entity any) *Accessor {
	typ := reflect.TypeOf(entity).Elem()
	numField := typ.NumField()
	fields := make(map[string]*FieldMeta, numField)
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		fields[field.Name] = &FieldMeta{
			Offset: field.Offset,
			typ:    field.Type,
		}
	}
	val := reflect.ValueOf(entity)
	return &Accessor{
		fields:  fields,
		address: val.UnsafePointer(),
	}
}

func (a *Accessor) Field(name string) (any, error) {
	meta, ok := a.fields[name]
	if !ok {
		return nil, errors.New("field not found")
	}
	fdAddr := unsafe.Pointer(uintptr(a.address) + meta.Offset)
	return reflect.NewAt(meta.typ, fdAddr).Elem().Interface(), nil
}

func (a *Accessor) SetField(name string, val any) error {
	meta, ok := a.fields[name]
	if !ok {
		return errors.New("field not found")
	}
	fdAddr := unsafe.Pointer(uintptr(a.address) + meta.Offset)
	reflect.NewAt(meta.typ, fdAddr).Elem().Set(reflect.ValueOf(val))
	return nil
}
