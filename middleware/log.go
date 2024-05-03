package middleware

import (
	"github.com/KNICEX/go-orm"
	"log"
)

type LogBuilder struct {
	logFunc func(query string, args []any)
}

func NewLogBuilder() *LogBuilder {
	return &LogBuilder{
		logFunc: func(query string, args []any) {
			log.Printf("query: %s, args: %v", query, args)
		},
	}
}

func (b *LogBuilder) LogFunc(f func(query string, args []any)) *LogBuilder {
	b.logFunc = f
	return b
}

func (b *LogBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx *orm.Context) *orm.Result {
			b.logFunc(ctx.Query.SQL, ctx.Query.Args)
			res := next(ctx)
			return res
		}
	}
}
