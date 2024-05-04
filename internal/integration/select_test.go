//go:build e2e

package integration

import (
	"context"
	"github.com/KNICEX/go-orm"
	"github.com/KNICEX/go-orm/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SelectSuite struct {
	Suite
}

func (s *SelectSuite) SetupSuite() {
	s.Suite.SetupSuite()
	err := orm.RawQuery[any](s.db, "TRUNCATE TABLE simple_struct;").Exec(context.Background()).Err()
	assert.NoError(s.T(), err)

	err = orm.NewInserter[test.SimpleStruct](s.db).
		Values(test.NewSimpleStruct(200)).Exec(context.Background()).Err()
	assert.NoError(s.T(), err)
}

func (s *SelectSuite) TearDownSuite() {
	err := orm.RawQuery[any](s.db, "TRUNCATE TABLE simple_struct;").Exec(context.Background()).Err()
	assert.NoError(s.T(), err)
}

func TestMySQLSelect(t *testing.T) {
	suite.Run(t, &SelectSuite{
		Suite{
			driver:  "mysql",
			dialect: orm.DialectMySQL,
			dsn:     "root:root@tcp(localhost:3307)/integration_test",
		},
	})
}

func (s *SelectSuite) TestSelect() {
	t := s.T()
	db := s.db
	testCases := []struct {
		name    string
		s       *orm.Selector[test.SimpleStruct]
		wantRes *test.SimpleStruct
		wantErr error
	}{
		{
			name:    "select one",
			s:       orm.NewSelector[test.SimpleStruct](db).Where(orm.Col("Id").Eq(200)),
			wantRes: test.NewSimpleStruct(200),
		},
		{
			name:    "no row",
			s:       orm.NewSelector[test.SimpleStruct](db).Where(orm.Col("Id").Eq(101)),
			wantErr: orm.ErrNoRows,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.s.Get(context.Background())
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
