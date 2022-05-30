package sql

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PGSSoft/terraform-provider-mssql/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestDatabaseTestSuite(t *testing.T) {
	s := &DatabaseTestSuite{}
	var (
		_ suite.AfterTest      = s
		_ suite.SetupTestSuite = s
	)
	suite.Run(t, s)
}

type DatabaseTestSuite struct {
	suite.Suite
	conn        connection
	db          database
	mock        sqlmock.Sqlmock
	diags       *diag.Diagnostics
	ctx         context.Context
	errExpected bool
}

func (s *DatabaseTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	s.Require().NoError(err, "SQL mock")
	s.mock = mock
	s.conn = connection{db: db}
	s.diags = &diag.Diagnostics{}
	s.ctx = utils.WithDiagnostics(context.Background(), s.diags)
	s.db = database{connection: s.conn, id: DatabaseId(rand.Int())}
}

func (s *DatabaseTestSuite) AfterTest(string, string) {
	if !s.errExpected {
		s.False(s.diags.HasError(), "Expected no errors in diagnostics, got: %v", s.diags)
	}

	s.NoError(s.mock.ExpectationsWereMet(), "SQL mock errors")
}

func (s *DatabaseTestSuite) TestGetDatabaseByName() {
	expectExactQuery(s.mock, "SELECT DB_ID(@p1)").WithArgs("test_db").WillReturnRows(newRows("ID").AddRow(21365))

	db := s.conn.GetDatabaseByName(s.ctx, "test_db")

	s.Equal(DatabaseId(21365), db.GetId(s.ctx), "DB ID")
}

func (s *DatabaseTestSuite) TestGetMultipleDatabases() {
	dbIds := []DatabaseId{1, 2}
	rows := newRows("database_id")
	for _, dbId := range dbIds {
		rows.AddRow(dbId)
	}
	s.expectDatabasesQuery().WillReturnRows(rows)

	dbs := s.conn.GetDatabases(s.ctx)

	s.Equal(2, len(dbs), "DBs count")

	for _, dbId := range dbIds {
		db, ok := dbs[dbId]
		s.True(ok, "DB with ID %d not found", dbId)
		s.Equal(dbId, db.(*database).id, "DB instance points to invalid ID")
	}
}

func (s *DatabaseTestSuite) TestGetDatabasesNoRows() {
	s.expectDatabasesQuery().WillReturnError(sql.ErrNoRows)

	s.Equal(0, len(s.conn.GetDatabases(s.ctx)), "Expected empty DB slice")
}

func (s *DatabaseTestSuite) TestGetDatabasesError() {
	err := errors.New("test_error")
	s.expectDatabasesQuery().WillReturnError(err)

	s.conn.GetDatabases(s.ctx)

	s.verifyError(err)
}

func (s *DatabaseTestSuite) TestExistsNoRows() {
	s.expectDatabaseSettingQuery().WillReturnError(sql.ErrNoRows)

	s.False(s.db.Exists(s.ctx))
}

func (s *DatabaseTestSuite) TestExistsSingleRow() {
	s.expectDatabaseSettingQuery().WillReturnRows(newRows("name", "collation_name").AddRow("name", "collation"))

	s.True(s.db.Exists(s.ctx))
}

func (s *DatabaseTestSuite) TestExistsError() {
	err := errors.New("test_error")
	s.expectDatabaseSettingQuery().WillReturnError(err)

	s.db.Exists(s.ctx)

	s.verifyError(err)
}

func (s *DatabaseTestSuite) TestCreteDatabaseNoCollation() {
	settings := DatabaseSettings{Name: "new_test_db"}
	dbId := DatabaseId(124)
	expectExactExec(s.mock, "CREATE DATABASE [%s]", settings.Name).WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectDatabaseIdQuery().WithArgs(settings.Name).WillReturnRows(newRows("ID").AddRow(dbId))

	db := s.conn.CreateDatabase(s.ctx, settings)

	s.Equal(dbId, db.GetId(s.ctx), "DB ID")
}

func (s *DatabaseTestSuite) TestCreteDatabaseWithCollation() {
	settings := DatabaseSettings{Name: "new_test_db", Collation: "new_test_db_collation"}
	dbId := DatabaseId(1223464)
	expectExactExec(s.mock, "CREATE DATABASE [%s] COLLATE %s", settings.Name, settings.Collation).
		WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectDatabaseIdQuery().WithArgs(settings.Name).WillReturnRows(newRows("ID").AddRow(dbId))

	db := s.conn.CreateDatabase(s.ctx, settings)

	s.Equal(dbId, db.GetId(s.ctx), "DB ID")
}

func (s *DatabaseTestSuite) TestGetSettings() {
	expSettings := DatabaseSettings{Name: "test_db_name", Collation: "test_collation"}
	s.expectDatabaseSettingQuery().
		WithArgs(s.db.id).
		WillReturnRows(newRows("name", "collation_name").AddRow(expSettings.Name, expSettings.Collation))

	actSettings := s.db.GetSettings(s.ctx)

	s.EqualValues(expSettings, actSettings)
}

func (s *DatabaseTestSuite) TestGetSettingsError() {
	err := errors.New("test_error")
	s.expectDatabaseSettingQuery().WithArgs(s.db.id).WillReturnError(err)

	s.db.GetSettings(s.ctx)

	s.verifyError(err)
}

func (s *DatabaseTestSuite) TestRename() {
	oldSettings := DatabaseSettings{Name: "old_db_name"}
	const newName = "new_db_name"
	s.expectDatabaseSettingQuery().WithArgs(s.db.id).WillReturnRows(newRows("name", "collation_name").AddRow(oldSettings.Name, oldSettings.Collation))
	expectExactExec(s.mock, "ALTER DATABASE [%s] MODIFY NAME = %s", oldSettings.Name, newName).WillReturnResult(sqlmock.NewResult(0, 1))

	s.db.Rename(s.ctx, newName)
}

func (s *DatabaseTestSuite) TestSetCollation() {
	const dbName = "test_db_name"
	const newCollation = "test_db_new_collation"
	s.expectDatabaseSettingQuery().WithArgs(s.db.id).WillReturnRows(newRows("name", "collation_name").AddRow(dbName, ""))
	expectExactExec(s.mock, "ALTER DATABASE [%s] COLLATE %s", dbName, newCollation).WillReturnResult(sqlmock.NewResult(0, 1))

	s.db.SetCollation(s.ctx, newCollation)
}

func (s *DatabaseTestSuite) TestDrop() {
	const dbName = "test_db_name"
	s.expectDatabaseSettingQuery().WithArgs(s.db.id).WillReturnRows(newRows("name", "collation_name").AddRow(dbName, ""))
	expectExactExec(s.mock, "DROP DATABASE [%s]", dbName).WillReturnResult(sqlmock.NewResult(0, 1))

	s.db.Drop(s.ctx)
}

func (s *DatabaseTestSuite) verifyError(err error) {
	s.errExpected = true

	for _, d := range *s.diags {
		if d.Severity() == diag.SeverityError && d.Detail() == err.Error() {
			return
		}
	}

	s.Failf("Missing error", "Could not find error '%s' in diagnostics. Full diagnostics: %v", err, s.diags)
}

func (s *DatabaseTestSuite) expectDatabasesQuery() *sqlmock.ExpectedQuery {
	return expectExactQuery(s.mock, "SELECT [database_id] FROM sys.databases")
}

func (s *DatabaseTestSuite) expectDatabaseSettingQuery() *sqlmock.ExpectedQuery {
	return expectExactQuery(s.mock, "SELECT [name], collation_name FROM sys.databases WHERE [database_id] = @p1").WithArgs(s.db.id)
}

func (s *DatabaseTestSuite) expectDatabaseIdQuery() *sqlmock.ExpectedQuery {
	return expectExactQuery(s.mock, "SELECT DB_ID(@p1)")
}
