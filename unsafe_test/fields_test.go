package unsafe_test

import (
	"reflect"
	"testing"
)

func PrintFieldOffset(entity any) {
	typ := reflect.TypeOf(entity)
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		println(field.Name, field.Offset)
	}
}

type User struct {
	Name    string
	Age     int
	Alias   []string
	Address string
}

func TestPrintFieldOffset(t *testing.T) {
	testCases := []struct {
		name   string
		entity any
	}{
		{
			name:   "test User",
			entity: User{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			PrintFieldOffset(tc.entity)
		})
	}
}
