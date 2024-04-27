package errs

import (
	"errors"
	"fmt"
)

var (
	ErrModelType = errors.New("orm: only support struct pointer")
)

func NewErrUnsupportedExpression(expr any) error {
	return fmt.Errorf("orm: unsupported expression type %v", expr)
}

func NewErrUnknownField(name string) error {
	return fmt.Errorf("orm: unknown field %s", name)
}

func NewErrInvalidTag(tag string) error {
	return fmt.Errorf("orm: invalid tag %s", tag)
}
