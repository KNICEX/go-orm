package errs

import (
	"errors"
	"fmt"
)

var (
	ErrModelType = errors.New("orm: only support struct pointer")
	ErrNoRows    = errors.New("orm: no rows")
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

func NewErrUnknownColumn(name string) error {
	return fmt.Errorf("orm: unknown column %s", name)
}
