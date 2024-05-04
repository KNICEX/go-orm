package unsafe

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccessor_Field(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	accessor := NewAccessor(&User{Name: "test", Age: 18})
	name, err := accessor.Field("name")
	require.NoError(t, err)
	require.Equal(t, "test", name)

}
