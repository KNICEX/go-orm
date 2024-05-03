package orm

import (
	"context"
	"github.com/KNICEX/go-orm/model"
)

const (
	SELECT = "SELECT"
	INSERT = "INSERT"
	UPDATE = "UPDATE"
	DELETE = "DELETE"
)

type Context struct {
	// 操作类型 INSERT, UPDATE, DELETE, SELECT
	Type string

	Query *Query

	Model *model.Model

	Ctx context.Context
}

type Result struct {
	// 数据库执行结果， SELECT 为查询后的结构体指针或者结构体指针切片， INSERT, UPDATE, DELETE 为 ExecResult
	Res any
	// 错误信息
	Err error
}

// Middleware 全局中间件
type Middleware func(next Handler) Handler

type Handler func(c *Context) *Result
