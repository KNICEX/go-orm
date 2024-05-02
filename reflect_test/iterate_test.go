package relect_test

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func IterateArrayOrSlice(entity any) ([]any, error) {
	val := reflect.ValueOf(entity)
	res := make([]any, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		ele := val.Index(i)
		res = append(res, ele.Interface())
	}
	return res, nil
}

func TestIterateArray(t *testing.T) {
	testCases := []struct {
		name     string
		entity   any
		wantVals []any
		wantErr  error
	}{
		{
			name:     "[]int",
			entity:   []int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
		{
			name:     "[3]string",
			entity:   [3]string{"a", "b", "c"},
			wantVals: []any{"a", "b", "c"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vals, err := IterateArrayOrSlice(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantVals, vals)
		})
	}
}

func IterateMap(entity any) ([]any, []any, error) {
	val := reflect.ValueOf(entity)
	resKeys := make([]any, 0, val.Len())
	resVals := make([]any, 0, val.Len())
	keys := val.MapKeys()
	for _, key := range keys {
		v := val.MapIndex(key)
		resKeys = append(resKeys, key.Interface())
		resVals = append(resVals, v.Interface())
	}
	return resKeys, resVals, nil
}

func TestIterateMap(t *testing.T) {
	testCases := []struct {
		name     string
		entity   any
		wantKeys []any
		wantVals []any
		wantErr  error
	}{
		{
			name:     "map[int]string",
			entity:   map[int]string{1: "a", 2: "b"},
			wantKeys: []any{1, 2},
			wantVals: []any{"a", "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keys, vals, err := IterateMap(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantKeys, keys)
			assert.Equal(t, tc.wantVals, vals)

		})
	}
}
