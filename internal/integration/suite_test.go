//go:build e2e

package integration

import (
	"github.com/KNICEX/go-orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	driver  string
	dsn     string
	dialect orm.Dialect
	db      *orm.DB
}

func (s *Suite) SetupSuite() {
	db, err := orm.Open(s.driver, s.dsn, orm.DBWithDialect(s.dialect))
	assert.NoError(s.T(), err)
	db.Wait()
	s.db = db
}
