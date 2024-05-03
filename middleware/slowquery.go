package middleware

import (
	"github.com/KNICEX/go-orm"
	"log"
	"time"
)

type SlowQueryBuilder struct {
	threshold time.Duration
	logFunc   func(query string, args []any, duration time.Duration)
}

func NewSlowQueryBuilder(threshold time.Duration) *SlowQueryBuilder {
	return &SlowQueryBuilder{
		threshold: threshold,
		logFunc: func(query string, args []any, duration time.Duration) {
			log.Printf("slow query: %s, args: %v, time: %v", query, args, duration)
		},
	}
}

func (b *SlowQueryBuilder) LogFunc(f func(query string, args []any, duration time.Duration)) *SlowQueryBuilder {
	b.logFunc = f
	return b
}

func (b *SlowQueryBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx *orm.Context) *orm.Result {
			start := time.Now()
			defer func() {
				d := time.Since(start)
				if d > b.threshold {
					b.logFunc(ctx.Query.SQL, ctx.Query.Args, d)
				}
			}()
			return next(ctx)
		}
	}
}
