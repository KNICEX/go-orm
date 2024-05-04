package errs

import (
	"errors"
	"fmt"
)

var (
	ErrModelType     = errors.New("orm: only support struct pointer")
	ErrNoRows        = errors.New("orm: no rows")
	ErrInsertZeroRow = errors.New("orm: insert zero row")
	ErrUpdateNoSet   = errors.New("orm: update no set")
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

func NewErrUnsupportedAssignable(assign any) error {
	return fmt.Errorf("orm: unsupported assignable type %v", assign)
}

func NewErrUnsupportedSetAble(setAble any) error {
	return fmt.Errorf("orm: unsupported setAble type %v", setAble)
}

func NewErrFailedToRollback(bizErr error, rollbackErr error) error {
	return fmt.Errorf("orm: failed to rollback transaction, business error: %w, rollback error: %v", bizErr, rollbackErr)
}

func NewErrTableNotExist(tableName string) error {
	return fmt.Errorf("orm: table %s does not exist", tableName)
}

func NewErrTableExist(tableName string) error {
	return fmt.Errorf("orm: table %s already exists", tableName)
}
