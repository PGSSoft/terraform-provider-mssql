package sql

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
)

func TestServerRoleTestSuite(t *testing.T) {
	s := &ServerRoleTestSuite{}
	suite.Run(t, s)
}

type ServerRoleTestSuite struct {
	SqlTestSuite
	role serverRole
}

func (s *ServerRoleTestSuite) SetupTest() {
	s.SqlTestSuite.SetupTest()
	s.role = serverRole{conn: s.connMock, id: ServerRoleId(rand.Int())}
}

func (s *ServerRoleTestSuite) TestGetServerRole() {
	role := GetServerRole(s.ctx, s.connMock, 135)

	s.Equal(ServerRoleId(135), role.GetId(s.ctx))
}

func (s *ServerRoleTestSuite) TestCreateServerRole() {
	s.expectServerPrincipalNameLookupQuery(245, "test_owner")
	expectExactExec(s.mock, "CREATE SERVER ROLE [test_role] AUTHORIZATION [test_owner]").WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectServerPrincipalIdLookupQuery(2457, "test_role")

	role := CreateServerRole(s.ctx, s.connMock, ServerRoleSettings{
		Name:    "test_role",
		OwnerId: 245,
	})

	s.Equal(ServerRoleId(2457), role.GetId(s.ctx))
}

func (s *ServerRoleTestSuite) TestCreateServerRoleDefaultOwner() {
	expectExactExec(s.mock, "CREATE SERVER ROLE [test_role]").WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectServerPrincipalIdLookupQuery(57, "test_role")

	role := CreateServerRole(s.ctx, s.connMock, ServerRoleSettings{
		Name:    "test_role",
		OwnerId: EmptyServerPrincipalId,
	})

	s.Equal(ServerRoleId(57), role.GetId(s.ctx))
}

func (s *ServerRoleTestSuite) TestGetSettings() {
	expectExactQuery(s.mock, "SELECT [name], [owning_principal_id] FROM sys.server_principals WHERE [principal_id]=@p1").
		WithArgs(s.role.id).
		WillReturnRows(newRows("name", "owning_principal_id").AddRow("test_name", 24521))

	settings := s.role.GetSettings(s.ctx)

	s.Equal(ServerRoleSettings{Name: "test_name", OwnerId: 24521}, settings)
}

func (s *ServerRoleTestSuite) TestRename() {
	s.expectServerPrincipalNameLookupQuery(int(s.role.id), "old_name")
	expectExactExec(s.mock, "ALTER SERVER ROLE [old_name] WITH NAME = [new_name]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.role.Rename(s.ctx, "new_name")
}

func (s *ServerRoleTestSuite) TestDrop() {
	s.expectServerPrincipalNameLookupQuery(int(s.role.id), "test_role")
	expectExactExec(s.mock, "DROP SERVER ROLE [test_role]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.role.Drop(s.ctx)
}
