package model

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...Option) (*Model, error)
}

func NewRegistry() Registry {
	return &registry{
		models: make(map[reflect.Type]*Model),
	}
}

type registry struct {
	lock   sync.RWMutex
	models map[reflect.Type]*Model
}

// Get 只接收结构体一级指针
func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)

	r.lock.RLock()
	m, ok := r.models[typ]
	r.lock.RUnlock()
	if ok {
		return m, nil
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	// double check
	m, ok = r.models[typ]
	if ok {
		return m, nil
	}

	var err error
	m, err = r.Register(val)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Register 只接受 struct 一级指针
func (r *registry) Register(entity any, opts ...Option) (*Model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Pointer {
		return nil, errs.ErrModelType
	}
	typ = typ.Elem()
	if typ.Kind() != reflect.Struct {
		return nil, errs.ErrModelType
	}

	numField := typ.NumField()
	fieldMap := make(map[string]*Field)
	colMap := make(map[string]*Field)
	fields := make([]*Field, 0, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		if fd.IsExported() {
			tags, err := r.parseTag(fd.Tag)
			if err != nil {
				return nil, err
			}

			colName := tags[tagColumn]
			if colName == "" {
				colName = underscoreName(fd.Name)
			}

			fieldInfo := &Field{
				ColName: colName,
				Typ:     fd.Type,
				GoName:  fd.Name,
				Offset:  fd.Offset,
			}
			fieldMap[fd.Name] = fieldInfo
			colMap[colName] = fieldInfo
			fields = append(fields, fieldInfo)
		}
	}

	var tableName string
	if tbn, ok := entity.(TableName); ok {
		tableName = tbn.TableName()
	} else {
		tableName = underscoreName(typ.Name())
	}

	res := &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		ColMap:    colMap,
		Fields:    fields,
	}
	for _, opt := range opts {
		if err := opt(res); err != nil {
			return nil, err
		}
	}
	r.models[typ] = res
	return res, nil
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag, ok := tag.Lookup("orm")
	if !ok || ormTag == "" {
		return map[string]string{}, nil
	}
	res := make(map[string]string)
	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		segs := strings.Split(pair, "=")
		if len(segs) == 1 {
			res[strings.TrimSpace(segs[0])] = ""
		}
		if len(segs) == 2 {
			key := strings.TrimSpace(segs[0])
			val := strings.TrimSpace(segs[1])
			res[key] = val
		}
	}
	return res, nil
}

func underscoreName(name string) string {
	var res []rune
	for i, v := range name {
		if unicode.IsUpper(v) {
			if i != 0 {
				res = append(res, '_')
			}
			res = append(res, unicode.ToLower(v))
		} else {
			res = append(res, v)
		}
	}
	return string(res)
}
