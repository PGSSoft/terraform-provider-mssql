package sql

import (
	"context"
	"database/sql"
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
	conn        connection
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
	s.conn = connection{conn: db, connDetails: ConnectionDetails{Auth: ConnectionAuthSql{}}}
	s.dbMock = dbMock{}
	s.dbMock.On("connect", mock.Anything).Return(db)
	s.dbMock.On("GetConnection", mock.Anything).Return(s.conn)
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
	return expectExactQuery(s.mock, "SELECT SUSER_SNAME(CONVERT(VARBINARY(85), @p1, 1))")
}

var _ Database = &dbMock{}

type dbMock struct {
	mock.Mock
}

func (m *dbMock) GetConnection(ctx context.Context) Connection {
	return m.Called(ctx).Get(0).(Connection)
}

func (m *dbMock) GetId(ctx context.Context) DatabaseId {
	return m.Called(ctx).Get(0).(DatabaseId)
}

func (m *dbMock) Exists(ctx context.Context) bool {
	return m.Called(ctx).Bool(0)
}

func (m *dbMock) GetSettings(ctx context.Context) DatabaseSettings {
	return m.Called(ctx).Get(0).(DatabaseSettings)
}

func (m *dbMock) Rename(ctx context.Context, name string) {
	m.Called(ctx, name)
}

func (m *dbMock) SetCollation(ctx context.Context, collation string) {
	m.Called(ctx, collation)
}

func (m *dbMock) Drop(ctx context.Context) {
	m.Called(ctx)
}

func (m *dbMock) CreateUser(ctx context.Context, settings UserSettings) User {
	return m.Called(ctx, settings).Get(0).(User)
}

func (m *dbMock) GetUser(ctx context.Context, id UserId) User {
	return m.Called(ctx, id).Get(0).(User)
}

func (m dbMock) connect(ctx context.Context) *sql.DB {
	return m.Called(ctx).Get(0).(*sql.DB)
}
