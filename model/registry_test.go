package model

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type TestModel struct {
	Id        int64
	FirstName string
	LastName  string
}

func TestRegistry_Get(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *Model
		fields    []*Field
		wantErr   error
	}{
		{
			name: "test Model",
			entity: TestModel{
				Id:        1,
				FirstName: "foo",
				LastName:  "bar",
			},
			wantErr: errs.ErrModelType,
		},
		{
			name: "test Model pointer",
			entity: &TestModel{
				Id: 1,
			},
			wantModel: &Model{
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Typ:     reflect.TypeOf(int64(0)),
					Offset:  0,
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Typ:     reflect.TypeOf(""),
					Offset:  24,
				},
			},
		},
		{
			name:    "not struct",
			entity:  1,
			wantErr: errs.ErrModelType,
		},
		{
			name:   "tag",
			entity: &TestModelWithTag{},
			wantModel: &Model{
				TableName: "test_model_with_tag",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Typ:     reflect.TypeOf(int64(0)),
					Offset:  0,
				},
				{
					ColName: "first_name_t",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "last_name_t",
					GoName:  "LastName",
					Typ:     reflect.TypeOf(""),
					Offset:  24,
				},
			},
		},
		{
			name:   "custom table name implement ",
			entity: &TestModelCustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name",
			},
			fields: []*Field{},
		},
		{
			name:   "custom table name pointer implement",
			entity: &TestModelCustomTableNamePointer{},
			wantModel: &Model{
				TableName: "custom_table_name_pointer",
			},
			fields: []*Field{},
		},
	}

	r := &registry{
		models: make(map[reflect.Type]*Model, 24),
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field)
			colMap := make(map[string]*Field)
			for _, f := range tc.fields {
				fieldMap[f.GoName] = f
				colMap[f.ColName] = f
			}
			tc.wantModel.Fields = tc.fields
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColMap = colMap
			assert.EqualValues(t, tc.wantModel, m)
		})
	}
}

type TestModelWithTag struct {
	Id        int64
	FirstName string `orm:"column=first_name_t"`
	LastName  string `orm:"column=last_name_t"`
}

type TestModelCustomTableName struct {
}

func (t TestModelCustomTableName) TableName() string {
	return "custom_table_name"
}

type TestModelCustomTableNamePointer struct {
}

func (t *TestModelCustomTableNamePointer) TableName() string {
	return "custom_table_name_pointer"
}

func TestRegistry_Register(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		opts      []Option
		fields    []*Field
		wantModel *Model
		wantErr   error
	}{
		{
			name:   "test Model",
			entity: &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Typ:     reflect.TypeOf(int64(0)),
					Offset:  0,
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Typ:     reflect.TypeOf(""),
					Offset:  24,
				},
			},
		},
		{
			name:   "test Model with TableNameOption",
			entity: &TestModel{},
			opts: []Option{
				WithTableName("test"),
			},
			wantModel: &Model{
				TableName: "test",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Typ:     reflect.TypeOf(int64(0)),
					Offset:  0,
				},
				{
					ColName: "first_name",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "last_name",
					GoName:  "LastName",
					Typ:     reflect.TypeOf(""),
					Offset:  24,
				},
			},
		},
		{
			name:   "test Model with ColNameOption",
			entity: &TestModel{},
			opts: []Option{
				WithColName("FirstName", "first_name_t"),
				WithColName("LastName", "last_name_t"),
			},
			wantModel: &Model{
				TableName: "test_model",
			},
			fields: []*Field{
				{
					ColName: "id",
					GoName:  "Id",
					Typ:     reflect.TypeOf(int64(0)),
					Offset:  0,
				},
				{
					ColName: "first_name_t",
					GoName:  "FirstName",
					Typ:     reflect.TypeOf(""),
					Offset:  8,
				},
				{
					ColName: "last_name_t",
					GoName:  "LastName",
					Typ:     reflect.TypeOf(""),
					Offset:  24,
				},
			},
		},
		{
			name:   "test Model with ColNameOption unknown field",
			entity: &TestModel{},
			opts: []Option{
				WithColName("Age", "user_age"),
			},
			wantErr: errs.NewErrUnknownField("Age"),
		},
	}

	r := &registry{
		models: make(map[reflect.Type]*Model, 24),
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Register(tc.entity, tc.opts...)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field)
			colMap := make(map[string]*Field)
			for _, f := range tc.fields {
				fieldMap[f.GoName] = f
				colMap[f.ColName] = f
			}
			tc.wantModel.Fields = tc.fields
			tc.wantModel.FieldMap = fieldMap
			tc.wantModel.ColMap = colMap
			assert.EqualValues(t, tc.wantModel, m)
		})
	}
}

type User struct {
	ID int `geeorm:"column=id"`
}
