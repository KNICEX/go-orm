package relect

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type FuncInfo struct {
	Name        string
	InputTypes  []reflect.Type
	OutputTypes []reflect.Type
	Result      []any
}

func IterateFunc(entity any) (map[string]FuncInfo, error) {
	if entity == nil {
		return nil, errors.New("nil value")
	}
	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)

	res := make(map[string]FuncInfo)

	numMethod := typ.NumMethod()
	for i := 0; i < numMethod; i++ {
		method := typ.Method(i)
		fn := method.Func
		numIn := fn.Type().NumIn()
		numOut := fn.Type().NumOut()
		inputTypes := make([]reflect.Type, numIn)
		inputValues := make([]reflect.Value, numIn)
		outputTypes := make([]reflect.Type, numOut)

		inputTypes[0] = fn.Type().In(0)
		inputValues[0] = val

		for j := 1; j < numIn; j++ {
			inputTypes[j] = fn.Type().In(j)
			inputValues[j] = reflect.Zero(fn.Type().In(j))
		}
		for j := 0; j < numOut; j++ {
			outputTypes[j] = fn.Type().Out(j)
		}
		resValues := fn.Call(inputValues)
		result := make([]any, numOut)
		for j := 0; j < numOut; j++ {
			result[j] = resValues[j].Interface()
		}
		res[method.Name] = FuncInfo{
			Name:        method.Name,
			InputTypes:  inputTypes,
			OutputTypes: outputTypes,
			Result:      result,
		}
	}

	return res, nil
}

func TestIterateFunc(t *testing.T) {
	testCases := []struct {
		name    string
		entity  any
		wantErr error
		wantRes map[string]FuncInfo
	}{
		{
			name: "user",
			entity: User{
				Id:   1,
				Name: "foo",
			},
			wantRes: map[string]FuncInfo{
				"GetName": {
					Name:        "GetName",
					InputTypes:  []reflect.Type{reflect.TypeOf(User{})},
					OutputTypes: []reflect.Type{reflect.TypeOf("")},
					Result:      []any{"foo"},
				},
			},
		},
		{
			name: "user pointer",
			entity: &User{
				Id:   1,
				Name: "foo",
			},
			wantRes: map[string]FuncInfo{
				"GetId": {
					Name:        "GetId",
					InputTypes:  []reflect.Type{reflect.TypeOf(&User{})},
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{1},
				},
				"GetName": {
					Name:        "GetName",
					InputTypes:  []reflect.Type{reflect.TypeOf(&User{})},
					OutputTypes: []reflect.Type{reflect.TypeOf("")},
					Result:      []any{"foo"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateFunc(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

type User struct {
	Id   int
	Name string
}

func (u User) GetName() string {
	return u.Name
}

func (u *User) GetId() int {
	return u.Id
}
