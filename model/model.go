package model

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"reflect"
)

const (
	tagColumn = "column"
)

type Model struct {
	typ       reflect.Type
	TableName string
	// 字段名 -> 字段信息
	FieldMap map[string]*Field
	// 列名 -> 字段信息
	ColMap map[string]*Field
	Fields []*Field

	// 分表键
	Sks map[string]struct{}
	Sf  ShardingFunc
}

// ShardingFunc 分表函数
type ShardingFunc func(sk map[string]any) (database string, table string)

type Option func(*Model) error

func WithTableName(name string) Option {
	return func(m *Model) error {
		m.TableName = name
		return nil
	}
}

func WithColName(field, colName string) Option {
	return func(m *Model) error {
		if _, ok := m.FieldMap[field]; !ok {
			return errs.NewErrUnknownField(field)
		}
		fieldInfo := m.FieldMap[field]
		// 移除旧的列名
		oldColName := fieldInfo.ColName
		delete(m.ColMap, oldColName)

		// 更新新的列名
		fieldInfo.ColName = colName
		m.ColMap[colName] = fieldInfo
		return nil
	}
}

type Field struct {
	ColName string
	// 代码中的字段名
	GoName string
	Typ    reflect.Type
	Offset uintptr
}

type TableName interface {
	TableName() string
}
