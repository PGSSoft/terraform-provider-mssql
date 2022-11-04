package sql

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/mock"
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
	connMock    *connectionMock
	dbMock      dbMock
	mock        sqlmock.Sqlmock
	diags       *diag.Diagnostics
	ctx         context.Context
	errExpected bool
}

func (s *SqlTestSuite) SetupTest() {
	db, sMock, err := sqlmock.New()
	s.Require().NoError(err, "SQL mock")
	s.mock = sMock
	s.connMock = &connectionMock{db: db}
	s.dbMock = dbMock{}
	s.dbMock.On("connect", mock.Anything).Return(db)
	s.dbMock.On("GetConnection", mock.Anything).Return(s.connMock)
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

func (s *SqlTestSuite) expectSqlLoginNameLookupQuery() *sqlmock.ExpectedQuery {
	return expectExactQuery(s.mock, "SELECT [name] FROM sys.sql_logins WHERE [sid]=CONVERT(VARBINARY(85), @p1, 1)")
}

func (s *SqlTestSuite) expectDatabasePrincipalIdLookupQuery(name string, id int) *sqlmock.ExpectedQuery {
	return expectExactQuery(s.mock, "SELECT DATABASE_PRINCIPAL_ID(@p1)").WithArgs(name).WillReturnRows(newRows("id").AddRow(id))
}

func (s *SqlTestSuite) expectUserNameQuery(id int, name string) {
	expectExactQuery(s.mock, "SELECT USER_NAME(@p1)").
		WithArgs(id).
		WillReturnRows(newRows("id").AddRow(name))
}

func (s *SqlTestSuite) expectEditionQuery(edition SQLEdition) {
	s.connMock.edition = edition
}

func (s *SqlTestSuite) expectServerPrincipalNameLookupQuery(id int, name string) {
	s.connMock.On("lookupServerPrincipalName", mock.Anything, GenericServerPrincipalId(id)).Return(name)
}

func (s *SqlTestSuite) expectServerPrincipalIdLookupQuery(id int, name string) {
	s.connMock.On("lookupServerPrincipalId", mock.Anything, name).Return(GenericServerPrincipalId(id))
}
