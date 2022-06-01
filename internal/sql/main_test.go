package sql

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"time"
)

var (
	_ suite.SetupTestSuite = &SqlTestSuite{}
	_ suite.AfterTest      = &SqlTestSuite{}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type SqlTestSuite struct {
	suite.Suite
	conn        connection
	mock        sqlmock.Sqlmock
	diags       *diag.Diagnostics
	ctx         context.Context
	errExpected bool
}

func (s *SqlTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err, "SQL mock")
	s.mock = mock
	s.conn = connection{db: db}
	s.diags = &diag.Diagnostics{}
	s.ctx = utils.WithDiagnostics(context.Background(), s.diags)
}

func (s *SqlTestSuite) AfterTest(string, string) {
	if !s.errExpected {
		s.False(s.diags.HasError(), "Expected no errors in diagnostics, got: %v", s.diags)
	}

	s.NoError(s.mock.ExpectationsWereMet(), "SQL mock errors")
}

func (s *SqlTestSuite) verifyError(err error) {
	s.errExpected = true

	for _, d := range *s.diags {
		if d.Severity() == diag.SeverityError && d.Detail() == err.Error() {
			return
		}
	}

	s.Failf("Missing error", "Could not find error '%s' in diagnostics. Full diagnostics: %v", err, s.diags)
}
