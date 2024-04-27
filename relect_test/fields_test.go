package relect

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func IterateFields(v any) (map[string]any, error) {
	if v == nil {
		return nil, errors.New("nil value")
	}
	fields := make(map[string]any)
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	if val.IsZero() {
		return nil, errors.New("nil value")
	}

	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errors.New("not a struct")
	}

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldType := typ.Field(i)
		fieldVal := val.Field(i)
		if fieldType.IsExported() {
			fields[fieldType.Name] = fieldVal.Interface()
		} else {
			fields[fieldType.Name] = reflect.Zero(fieldType.Type).Interface()
		}

	}

	return fields, nil
}

func TestIterateFields(t *testing.T) {

	type User struct {
		Id    int
		Name  string
		inner string
	}

	testCases := []struct {
		name    string
		entity  any
		wantErr error
		wantRes map[string]any
	}{
		{
			name:   "user",
			entity: User{Id: 1, Name: "foo", inner: "bar"},
			wantRes: map[string]any{
				"Id":    1,
				"Name":  "foo",
				"inner": "",
			},
		},
		{
			name: "user pointer",
			entity: &User{
				Id:    1,
				Name:  "foo",
				inner: "bar",
			},
			wantRes: map[string]any{
				"Id":    1,
				"Name":  "foo",
				"inner": "",
			},
		},
		{
			name:    "basic",
			entity:  19,
			wantErr: errors.New("not a struct"),
		},
		{
			name:    "nil",
			entity:  nil,
			wantErr: errors.New("nil value"),
		},
		{
			name:    "nil User",
			entity:  (*User)(nil),
			wantErr: errors.New("nil value"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateFields(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func SetField(entity any, field string, newVal any) error {
	val := reflect.ValueOf(entity)
	for val.Type().Kind() == reflect.Pointer {
		val = val.Elem()
	}
	fieldVal := val.FieldByName(field)
	if !fieldVal.CanSet() {
		return errors.New("can not set")
	}
	fieldVal.Set(reflect.ValueOf(newVal))
	return nil
}

func TestSetField(t *testing.T) {

	type User struct {
		Id    int
		Name  string
		inner string
	}

	testCases := []struct {
		name       string
		entity     any
		field      string
		newVal     any
		wantEntity any
		wantErr    error
	}{
		{
			name:       "user pointer",
			entity:     &User{Id: 1, Name: "foo", inner: "bar"},
			field:      "Name",
			newVal:     "bar",
			wantEntity: &User{Id: 1, Name: "bar", inner: "bar"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetField(tc.entity, tc.field, tc.newVal)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantEntity, tc.entity)
		})
	}
}
