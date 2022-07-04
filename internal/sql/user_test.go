package sql

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
)

func TestUserTestSuite(t *testing.T) {
	s := &UserTestSuite{}
	suite.Run(t, s)
}

type UserTestSuite struct {
	SqlTestSuite
	user user
}

func (s *UserTestSuite) SetupTest() {
	s.SqlTestSuite.SetupTest()
	s.user = user{db: &s.dbMock, id: UserId(rand.Int())}
}

func (s *UserTestSuite) TestCreateSqlUser() {
	settings := UserSettings{Name: "test_user", LoginId: "test_login_id"}
	s.expectSqlLoginNameLookupQuery().WithArgs("test_login_id").WillReturnRows(newRows("name").AddRow("test_login"))
	expectExactExec(s.mock, "CREATE USER [test_user] FOR LOGIN [test_login]").
		WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectUserIdLookupQuery("test_user", 123)

	user := CreateUser(s.ctx, &s.dbMock, settings)

	s.Equal(UserId(123), user.GetId(s.ctx))
}

func (s *UserTestSuite) TestGetSqlUserByName() {
	s.expectUserIdLookupQuery("test_user_by_name", 521)

	user := GetUserByName(s.ctx, &s.dbMock, "test_user_by_name")

	s.Equal(UserId(521), user.GetId(s.ctx))
}

func (s *UserTestSuite) TestGetUsers() {
	expectExactQuery(s.mock, "SELECT [principal_id] FROM sys.database_principals WHERE [type] = 'S' AND [sid] IS NOT NULL").
		WillReturnRows(newRows("id").AddRow(3).AddRow(145))

	users := GetUsers(s.ctx, &s.dbMock)

	for _, id := range []UserId{3, 145} {
		s.Contains(users, id, "contains user")
		s.Equal(id, users[id].GetId(s.ctx), "user ID")
	}
}

func (s *UserTestSuite) TestGetSettings() {
	expectExactQuery(s.mock, "SELECT [name], CONVERT(VARCHAR(85), [sid], 1) FROM sys.database_principals WHERE [principal_id]=@p1").
		WithArgs(s.user.id).
		WillReturnRows(newRows("name", "login_id").AddRow("test_name", "test_login_id"))

	settings := s.user.GetSettings(s.ctx)

	s.Equal("test_name", settings.Name)
	s.Equal(LoginId("test_login_id"), settings.LoginId)
}

func (s *UserTestSuite) TestDrop() {
	s.expectUserNameQuery(int(s.user.id), "test_drop_name")
	expectExactExec(s.mock, "DROP USER [test_drop_name]").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.user.Drop(s.ctx)
}

func (s *UserTestSuite) TestUpdateSettings() {
	newSettings := UserSettings{Name: "new_name", LoginId: "new_login_id"}
	s.expectUserNameQuery(int(s.user.id), "test_update_settings")
	s.expectSqlLoginNameLookupQuery().WithArgs(newSettings.LoginId).WillReturnRows(newRows("name").AddRow("new_login_name"))
	expectExactExec(s.mock, "ALTER USER [test_update_settings] WITH NAME=[%s], LOGIN=[%s]", newSettings.Name, "new_login_name").
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.user.UpdateSettings(s.ctx, newSettings)
}

func (s *UserTestSuite) expectUserIdLookupQuery(name string, id int) {
	expectExactQuery(s.mock, "SELECT USER_ID(@p1)").WithArgs(name).WillReturnRows(newRows("id").AddRow(id))
}
