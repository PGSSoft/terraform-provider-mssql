package sql

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
)

func TestDatabaseTestSuite(t *testing.T) {
	s := &DatabaseTestSuite{}
	suite.Run(t, s)
}

type DatabaseTestSuite struct {
	db database
	SqlTestSuite
}

func (s *DatabaseTestSuite) SetupTest() {
	s.SqlTestSuite.SetupTest()
	s.db = database{conn: s.conn, id: DatabaseId(rand.Int())}
}

func (s *DatabaseTestSuite) TestGetDatabaseByName() {
	s.expectDatabaseIdQuery().WillReturnRows(newRows("ID").AddRow(21365))

	db := GetDatabaseByName(s.ctx, s.conn, "test_db")

	s.Equal(DatabaseId(21365), db.GetId(s.ctx), "DB ID")
}

func (s *DatabaseTestSuite) TestGetMultipleDatabases() {
	dbIds := []DatabaseId{1, 2}
	rows := newRows("database_id")
	for _, dbId := range dbIds {
		rows.AddRow(dbId)
	}
	s.expectDatabasesQuery().WillReturnRows(rows)

	dbs := GetDatabases(s.ctx, s.conn)

	s.Equal(2, len(dbs), "DBs count")

	for _, dbId := range dbIds {
		db, ok := dbs[dbId]
		s.True(ok, "DB with ID %d not found", dbId)
		s.Equal(dbId, db.(*database).id, "DB instance points to invalid ID")
	}
}

func (s *DatabaseTestSuite) TestGetDatabasesNoRows() {
	s.expectDatabasesQuery().WillReturnError(sql.ErrNoRows)

	s.Equal(0, len(GetDatabases(s.ctx, s.conn)), "Expected empty DB slice")
}

func (s *DatabaseTestSuite) TestGetDatabasesError() {
	err := errors.New("test_error")
	s.expectDatabasesQuery().WillReturnError(err)

	GetDatabases(s.ctx, s.conn)

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

	db := CreateDatabase(s.ctx, s.conn, settings)

	s.Equal(dbId, db.GetId(s.ctx), "DB ID")
}

func (s *DatabaseTestSuite) TestCreteDatabaseWithCollation() {
	settings := DatabaseSettings{Name: "new_test_db", Collation: "new_test_db_collation"}
	dbId := DatabaseId(1223464)
	expectExactExec(s.mock, "CREATE DATABASE [%s] COLLATE %s", settings.Name, settings.Collation).
		WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectDatabaseIdQuery().WithArgs(settings.Name).WillReturnRows(newRows("ID").AddRow(dbId))

	db := CreateDatabase(s.ctx, s.conn, settings)

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

func (s *DatabaseTestSuite) expectDatabasesQuery() *sqlmock.ExpectedQuery {
	return expectExactQuery(s.mock, "SELECT [database_id] FROM sys.databases")
}

func (s *DatabaseTestSuite) expectDatabaseSettingQuery() *sqlmock.ExpectedQuery {
	return expectExactQuery(s.mock, "SELECT [name], collation_name FROM sys.databases WHERE [database_id] = @p1").WithArgs(s.db.id)
}

func (s *DatabaseTestSuite) expectDatabaseIdQuery() *sqlmock.ExpectedQuery {
	return expectExactQuery(s.mock, "SELECT database_id FROM sys.databases WHERE [name] = @p1")
}
