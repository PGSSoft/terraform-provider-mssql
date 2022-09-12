package sql

import (
	"math/rand"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
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
	settings := UserSettings{Name: "test_user", LoginId: "test_login_id", Type: USER_TYPE_SQL}
	s.expectSqlLoginNameLookupQuery().WithArgs("test_login_id").WillReturnRows(newRows("name").AddRow("test_login"))
	expectExactExec(s.mock, "CREATE USER [test_user] FOR LOGIN [test_login]").
		WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectUserIdLookupQuery("test_user", 123)

	user := CreateUser(s.ctx, &s.dbMock, settings)

	s.Equal(UserId(123), user.GetId(s.ctx))
}

func (s *UserTestSuite) TestCreateAzureADUser() {
	settings := UserSettings{Name: "test_user", AADObjectId: "e86c631e-8e80-46ab-a82f-04f11ec5740e", Type: USER_TYPE_AZUREAD}
	expectExactExec(s.mock, `
DECLARE @SQL NVARCHAR(MAX) = 'CREATE USER [' + @p1 + '] WITH SID=' + (SELECT CONVERT(VARCHAR(85), CONVERT(VARBINARY(85), CAST(@p2 AS UNIQUEIDENTIFIER), 1), 1)) + ', TYPE=E';
EXEC(@SQL)
`).WithArgs(settings.Name, settings.AADObjectId).WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectUserIdLookupQuery("test_user", 421)

	user := CreateUser(s.ctx, &s.dbMock, settings)

	s.Equal(UserId(421), user.GetId(s.ctx))
}

func (s *UserTestSuite) TestGetSqlUserByName() {
	s.expectUserIdLookupQuery("test_user_by_name", 521)

	user := GetUserByName(s.ctx, &s.dbMock, "test_user_by_name")

	s.Equal(UserId(521), user.GetId(s.ctx))
}

func (s *UserTestSuite) TestGetUsers() {
	expectExactQuery(s.mock, "SELECT [principal_id] FROM sys.database_principals WHERE [type] IN ('S', 'E', 'X') AND [sid] IS NOT NULL").
		WillReturnRows(newRows("id").AddRow(3).AddRow(145))

	users := GetUsers(s.ctx, &s.dbMock)

	for _, id := range []UserId{3, 145} {
		s.Contains(users, id, "contains user")
		s.Equal(id, users[id].GetId(s.ctx), "user ID")
	}
}

func (s *UserTestSuite) TestGetSettings() {
	s.expectSettingsQuery("S")

	settings := s.user.GetSettings(s.ctx)

	s.Equal("test_name", settings.Name)
	s.Equal(LoginId("test_login_id"), settings.LoginId)
	s.Equal(USER_TYPE_SQL, settings.Type, "type")
	s.Equal(AADObjectId(""), settings.AADObjectId, "object_id")
}

func (s *UserTestSuite) TestGetSettingsAzureAD() {
	s.expectSettingsQuery("E")

	settings := s.user.GetSettings(s.ctx)

	s.Equal("test_name", settings.Name)
	s.Equal(LoginId("test_login_id"), settings.LoginId)
	s.Equal(USER_TYPE_AZUREAD, settings.Type, "type")
	s.Equal(AADObjectId("67f1ec25-847b-4440-98c0-26dc0ad9d1f0"), settings.AADObjectId, "object_id")
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

func (s *UserTestSuite) expectSettingsQuery(userType string) {
	expectExactQuery(s.mock, "SELECT [name], CONVERT(VARCHAR(85), [sid], 1), [type], CONVERT(VARCHAR(36), CONVERT(UNIQUEIDENTIFIER, [sid], 1), 1) FROM sys.database_principals WHERE [principal_id]=@p1").
		WithArgs(s.user.id).
		WillReturnRows(newRows("name", "login_id", "type", "object_id").AddRow("test_name", "test_login_id", userType, "67f1ec25-847b-4440-98c0-26dc0ad9d1f0"))

}
