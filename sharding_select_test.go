package orm

import (
	"fmt"
	"github.com/KNICEX/go-orm/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

type Order struct {
	UserId int64
}

func TestShardingSelector_findDstByPredicate(t *testing.T) {
	testCases := []struct {
		name     string
		p        Predicate
		wantDsts []Dst
		wantErr  error
	}{
		{
			name: "only equal",
			p:    Col("UserId").Eq(11),
			wantDsts: []Dst{
				{
					Database: "order_db_0",
					Table:    "order_table_1",
				},
			},
		},
		{
			name: "eq and eq",
			p:    Col("UserId").Eq(222).And(Col("UserId").Eq(435)),
			wantDsts: []Dst{
				{
					Database: "order_db_2",
					Table:    "order_table_2",
				},
				{
					Database: "order_db_4",
					Table:    "order_table_5",
				},
			},
		},
	}
	r := model.NewRegistry()
	m, err := r.Get(&Order{})
	require.NoError(t, err)
	m.Sf = func(sk map[string]any) (database string, table string) {
		userId, ok := sk["UserId"]
		if !ok {
			return "", ""
		}
		uid, err := strconv.ParseInt(fmt.Sprintf("%v", userId), 10, 64)
		if err != nil {
			panic(err)
		}
		db := uid / 100
		tbl := uid % 10
		return fmt.Sprintf("order_db_%d", db), fmt.Sprintf("order_table_%d", tbl)
	}
	m.Sks = map[string]struct{}{
		"UserId": {},
	}

	s := ShardingSelector[Order]{
		builder: builder{
			core: &core{
				model: m,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dsts, err := s.findDstByPredicate(tc.p)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantDsts, dsts)
		})
	}
}
