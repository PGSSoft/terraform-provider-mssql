package sql

import (
	"database/sql"
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

func (s *ServerRoleTestSuite) TestGetServerRoles() {
	expectExactQuery(s.mock, "SELECT [principal_id] FROM sys.server_principals WHERE [type]='R'").
		WillReturnRows(newRows("principal_id").AddRow(23).AddRow(63))

	roles := GetServerRoles(s.ctx, s.connMock)

	s.Len(roles, 2, "count")
	s.Require().Contains(roles, ServerRoleId(23))
	s.Equal(ServerRoleId(23), roles[23].GetId(s.ctx))
	s.Require().Contains(roles, ServerRoleId(63))
	s.Equal(ServerRoleId(63), roles[63].GetId(s.ctx))
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

func (s *ServerRoleTestSuite) TestAddMember() {
	s.expectServerPrincipalNameLookupQuery(int(s.role.id), "test_role")
	s.expectServerPrincipalNameLookupQuery(243, "test_member")
	expectExactExec(s.mock, "ALTER SERVER ROLE [test_role] ADD MEMBER [test_member]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.role.AddMember(s.ctx, 243)
}

func (s *ServerRoleTestSuite) TestHasMember() {
	cases := map[string]struct {
		rows     *sqlmock.Rows
		expected bool
	}{
		"true": {
			rows:     newRows("no_name").AddRow(1),
			expected: true,
		},
		"false": {
			rows:     nil,
			expected: false,
		},
	}

	for name, tc := range cases {
		testCase := tc
		s.Run(name, func() {
			exp := expectExactQuery(s.mock, "SELECT 1 FROM sys.server_role_members WHERE [role_principal_id]=@p1 AND [member_principal_id]=@p2").
				WithArgs(s.role.id, 56)

			if testCase.rows == nil {
				exp.WillReturnError(sql.ErrNoRows)
			} else {
				exp.WillReturnRows(testCase.rows)
			}

			result := s.role.HasMember(s.ctx, 56)

			s.Equal(testCase.expected, result)
		})

	}
}

func (s *ServerRoleTestSuite) TestRemoveMember() {
	s.expectServerPrincipalNameLookupQuery(int(s.role.id), "test_role")
	s.expectServerPrincipalNameLookupQuery(19, "test_member")
	expectExactExec(s.mock, "ALTER SERVER ROLE [test_role] DROP MEMBER [test_member]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.role.RemoveMember(s.ctx, 19)
}

func (s *ServerRoleTestSuite) TestGetMembers() {
	expectExactQuery(s.mock, `
SELECT [principal_id], [name], [type] FROM sys.server_role_members
INNER JOIN sys.server_principals ON [member_principal_id] = [principal_id]
WHERE [role_principal_id]=@p1 AND [type] IN ('S', 'R')`).
		WithArgs(s.role.id).
		WillReturnRows(newRows("principal_id", "name", "type").AddRow(24, "role_name", "R").AddRow(64, "login_name", "S").AddRow(13, "unknown_name", "C"))

	members := s.role.GetMembers(s.ctx)

	expected := ServerRoleMembers{
		24: {
			Id:   24,
			Name: "role_name",
			Type: SERVER_ROLE,
		},
		64: {
			Id:   64,
			Name: "login_name",
			Type: SQL_LOGIN,
		},
		13: {
			Id:   13,
			Name: "unknown_name",
			Type: UNKNOWN,
		},
	}
	s.Equal(expected, members)
}
