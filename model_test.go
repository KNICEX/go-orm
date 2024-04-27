package orm

import (
	"github.com/KNICEX/go-orm/internal/errs"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_parseModel(t *testing.T) {
	testCases := []struct {
		name      string
		entity    any
		wantModel *model
		wantErr   error
	}{
		{
			name: "test model",
			entity: TestModel{
				Id:        1,
				FirstName: "foo",
				LastName:  "bar",
			},
			wantErr: errs.ErrModelType,
		},
		{
			name: "test model pointer",
			entity: &TestModel{
				Id: 1,
			},
			wantModel: &model{
				tableName: "test_model",
				fields: map[string]*field{
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name",
					},
					"LastName": {
						colName: "last_name",
					},
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
			wantModel: &model{
				tableName: "test_model_with_tag",
				fields: map[string]*field{
					"Id": {
						colName: "id",
					},
					"FirstName": {
						colName: "first_name_t",
					},
					"LastName": {
						colName: "last_name_t",
					},
				},
			},
		},
		{
			name:   "custom table name implement ",
			entity: &TestModelCustomTableName{},
			wantModel: &model{
				tableName: "custom_table_name",
				fields:    make(map[string]*field),
			},
		},
		{
			name:   "custom table name pointer implement",
			entity: &TestModelCustomTableNamePointer{},
			wantModel: &model{
				tableName: "custom_table_name_pointer",
				fields:    make(map[string]*field),
			},
		},
	}

	r := &registry{
		models: make(map[reflect.Type]*model, 24),
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.get(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

type TestModelWithTag struct {
	Id        int
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
