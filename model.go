package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

const (
	tagColumn = "column"
)

type model struct {
	tableName string
	fields    map[string]*field
}

type field struct {
	colName string
}

type registry struct {
	lock   sync.RWMutex
	models map[reflect.Type]*model
}

// 只接收结构体一级指针
func (r *registry) get(val any) (*model, error) {
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
	m, err = r.parseModel(val)
	if err != nil {
		return nil, err
	}
	r.models[typ] = m
	return m, nil
}

// parseModel struct 一级指针
func (r *registry) parseModel(entity any) (*model, error) {
	typ := reflect.TypeOf(entity)
	if typ.Kind() != reflect.Pointer {
		return nil, errs.ErrModelType
	}
	typ = typ.Elem()
	if typ.Kind() != reflect.Struct {
		return nil, errs.ErrModelType
	}

	numField := typ.NumField()
	fieldMap := make(map[string]*field)
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

			fieldMap[fd.Name] = &field{
				colName: colName,
			}
		}
	}

	var tableName string
	if tbn, ok := entity.(TableName); ok {
		tableName = tbn.TableName()
	} else {
		tableName = underscoreName(typ.Name())
	}

	return &model{
		tableName: tableName,
		fields:    fieldMap,
	}, nil
}

type User struct {
	ID int `geeorm:"column=id"`
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
