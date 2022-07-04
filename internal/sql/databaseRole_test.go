package sql

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
)

func TestDatabaseRoleTestSuite(t *testing.T) {
	s := &DatabaseRoleTestSuite{}
	suite.Run(t, s)
}

type DatabaseRoleTestSuite struct {
	SqlTestSuite
	role databaseRole
}

func (s *DatabaseRoleTestSuite) SetupTest() {
	s.SqlTestSuite.SetupTest()
	s.role = databaseRole{db: &s.dbMock, id: DatabaseRoleId(rand.Int())}
}

func (s *DatabaseRoleTestSuite) TestCreateDatabaseRoleWithoutOwner() {
	expectExactExec(s.mock, "CREATE ROLE [test_role]").WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectDatabasePrincipalIdLookupQuery("test_role", int(s.role.id))

	role := CreateDatabaseRole(s.ctx, &s.dbMock, "test_role", EmptyDatabasePrincipalId)

	s.Equal(s.role.id, role.GetId(s.ctx), "role ID")
}

func (s *DatabaseRoleTestSuite) TestCreateDatabaseRoleWithOwner() {
	s.expectUserNameQuery(134, "owner")
	expectExactExec(s.mock, "CREATE ROLE [test_role] AUTHORIZATION [owner]").WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectDatabasePrincipalIdLookupQuery("test_role", int(s.role.id))

	role := CreateDatabaseRole(s.ctx, &s.dbMock, "test_role", GenericDatabasePrincipalId(134))

	s.Equal(s.role.id, role.GetId(s.ctx))
}

func (s *DatabaseRoleTestSuite) TestGetOwnerId() {
	expectExactQuery(s.mock, "SELECT owning_principal_id FROM sys.database_principals WHERE principal_id=@p1").
		WithArgs(s.role.id).
		WillReturnRows(newRows("owning_principal_id").AddRow(135))

	ownerId := s.role.GetOwnerId(s.ctx)

	s.Equal(GenericDatabasePrincipalId(135), ownerId, "owner ID")
}

func (s *DatabaseRoleTestSuite) TestGetName() {
	s.expectUserNameQuery(int(s.role.id), "test_name")

	name := s.role.GetName(s.ctx)

	s.Equal("test_name", name, "name")
}

func (s *DatabaseRoleTestSuite) TestDrop() {
	s.expectUserNameQuery(int(s.role.id), "test_role")
	expectExactExec(s.mock, "DROP ROLE [test_role]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.role.Drop(s.ctx)
}

func (s *DatabaseRoleTestSuite) TestRename() {
	s.expectUserNameQuery(int(s.role.id), "test_role")
	expectExactExec(s.mock, "ALTER ROLE [test_role] WITH NAME = [new_name]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.role.Rename(s.ctx, "new_name")
}

func (s *DatabaseRoleTestSuite) TestChangeOwner() {
	s.expectUserNameQuery(int(s.role.id), "test_role")
	s.expectUserNameQuery(358, "new_owner")
	expectExactExec(s.mock, "ALTER AUTHORIZATION ON ROLE::[test_role] TO [new_owner]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.role.ChangeOwner(s.ctx, GenericDatabasePrincipalId(358))
}

func (s *DatabaseRoleTestSuite) TestGetDatabaseRoles() {
	expectExactQuery(s.mock, "SELECT [principal_id] FROM sys.database_principals WHERE [type] = 'R'").
		WillReturnRows(newRows("principal_id").AddRow(24).AddRow(2145))

	roles := GetDatabaseRoles(s.ctx, &s.dbMock)

	s.Len(roles, 2, "Number of roles")
	for _, id := range []DatabaseRoleId{24, 2145} {
		s.Equal(id, roles[id].GetId(s.ctx), "Role ID")
	}
}
