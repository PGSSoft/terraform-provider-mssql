package sql

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestSchemaTestSuite(t *testing.T) {
	s := &SchemaTestSuite{}
	suite.Run(t, s)
}

type SchemaTestSuite struct {
	SqlTestSuite
	schema Schema
}

func (s *SchemaTestSuite) SetupTest() {
	s.SqlTestSuite.SetupTest()
	s.schema = GetSchema(s.ctx, &s.dbMock, 322)
}

func (s *SchemaTestSuite) TestGetSchemaByName() {
	expectExactQuery(s.mock, "SELECT SCHEMA_ID(@p1)").WithArgs("test_schema").WillReturnRows(newRows("id").AddRow(235))

	sch := GetSchemaByName(s.ctx, &s.dbMock, "test_schema")

	s.Equal(235, int(sch.GetId(s.ctx)), "id")
}

func (s *SchemaTestSuite) TestGetSchemaByNameNotExists() {
	expectExactQuery(s.mock, "SELECT SCHEMA_ID(@p1)").WithArgs("not_exists").WillReturnRows(newRows("id").AddRow(nil))

	GetSchemaByName(s.ctx, &s.dbMock, "not_exists")

	s.errExpected = true
	for _, d := range *s.diags {
		if d.Severity() == diag.SeverityError && strings.Contains(d.Summary(), "not exist") {
			return
		}
	}

	s.Fail("Did not found correct error")
}

func (s *SchemaTestSuite) TestGetSchemas() {
	expectExactQuery(s.mock, "SELECT [schema_id] FROM sys.schemas").WillReturnRows(newRows("schema_id").AddRow(1241).AddRow(543))

	schemas := GetSchemas(s.ctx, &s.dbMock)

	s.Require().Len(schemas, 2, "count")
	s.Require().Contains(schemas, SchemaId(1241))
	s.Require().Contains(schemas, SchemaId(543))
	s.Equal(SchemaId(1241), schemas[1241].GetId(s.ctx))
	s.Equal(SchemaId(543), schemas[543].GetId(s.ctx))
}

func (s *SchemaTestSuite) TestCreateSchemaWithDefaultOwner() {
	s.dbMock.On("getUserName", mock.Anything, EmptyDatabasePrincipalId).Return("self")
	expectExactExec(s.mock, "CREATE SCHEMA [test_schema] AUTHORIZATION [self]").WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectSchemaIdQuery("test_schema", 13)

	sch := CreateSchema(s.ctx, &s.dbMock, "test_schema", EmptyDatabasePrincipalId)

	s.Equal(13, int(sch.GetId(s.ctx)), "id")
}

func (s *SchemaTestSuite) TestCreateSchemaWithOwner() {
	s.dbMock.On("getUserName", mock.Anything, GenericDatabasePrincipalId(634)).Return("test_owner")
	expectExactExec(s.mock, "CREATE SCHEMA [test_schema_with_owner] AUTHORIZATION [test_owner]").WillReturnResult(sqlmock.NewResult(0, 1))
	s.expectSchemaIdQuery("test_schema_with_owner", 24)

	sch := CreateSchema(s.ctx, &s.dbMock, "test_schema_with_owner", DatabaseRoleId(634))

	s.Equal(24, int(sch.GetId(s.ctx)), "id")
}

func (s *SchemaTestSuite) TestGetOwnerId() {
	expectExactQuery(s.mock, "SELECT [principal_id] FROM sys.schemas WHERE [schema_id] = @p1").
		WithArgs(s.schema.GetId(s.ctx)).
		WillReturnRows(newRows("principal_id").AddRow(425))

	ownerId := s.schema.GetOwnerId(s.ctx)

	s.Equal(425, int(ownerId), "owner id")
}

func (s *SchemaTestSuite) TestChangeOwner() {
	s.expectSchemaNameQuery("test_schema_chown", int(s.schema.GetId(s.ctx)))
	s.dbMock.On("getUserName", mock.Anything, GenericDatabasePrincipalId(23)).Return("new_owner")
	expectExactExec(s.mock, "ALTER AUTHORIZATION ON schema::[test_schema_chown] TO [new_owner]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.schema.ChangeOwner(s.ctx, 23)
}

func (s *SchemaTestSuite) TestChangeOwnerToCurrent() {
	s.expectSchemaNameQuery("test_schema_chown", int(s.schema.GetId(s.ctx)))
	s.dbMock.On("getUserName", mock.Anything, EmptyDatabasePrincipalId).Return("self")
	expectExactExec(s.mock, "ALTER AUTHORIZATION ON schema::[test_schema_chown] TO [self]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.schema.ChangeOwner(s.ctx, EmptyDatabasePrincipalId)
}

func (s *SchemaTestSuite) TestDrop() {
	s.expectSchemaNameQuery("to_be_dropped", int(s.schema.GetId(s.ctx)))
	expectExactExec(s.mock, "DROP SCHEMA [to_be_dropped]").WillReturnResult(sqlmock.NewResult(0, 1))

	s.schema.Drop(s.ctx)
}

func (s *SchemaTestSuite) TestGetPermissions() {
	expectExactQuery(s.mock, "SELECT [permission_name], [state] FROM sys.database_permissions WHERE [class]=3 AND [major_id]=@p1 AND [grantee_principal_id]=@p2").
		WithArgs(s.schema.GetId(s.ctx), 135).
		WillReturnRows(newRows("permission_name", "state").AddRow("TEST1", "W").AddRow("TEST2", "G"))

	perms := s.schema.GetPermissions(s.ctx, 135)

	s.Len(perms, 2, "count")
	s.Require().Contains(perms, "TEST1")
	s.Equal(SchemaPermission{Name: "TEST1", WithGrantOption: true}, perms["TEST1"])
	s.Require().Contains(perms, "TEST2")
	s.Equal(SchemaPermission{Name: "TEST2", WithGrantOption: false}, perms["TEST2"])
}

func (s *SchemaTestSuite) TestGrantPermission() {
	s.expectSchemaNameQuery("test_schema", int(s.schema.GetId(s.ctx)))
	s.dbMock.expectUsernameLookup(631, "test_user")
	expectExactExec(s.mock, "GRANT TEST_PERM ON schema::[test_schema] TO [test_user]")

	s.schema.GrantPermission(s.ctx, 631, SchemaPermission{Name: "TEST_PERM"})
}

func (s *SchemaTestSuite) TestGrantPermissionWithGrantOption() {
	s.expectSchemaNameQuery("test_schema", int(s.schema.GetId(s.ctx)))
	s.dbMock.expectUsernameLookup(151, "test_user")
	expectExactExec(s.mock, "GRANT TEST_PERM2 ON schema::[test_schema] TO [test_user] WITH GRANT OPTION")

	s.schema.GrantPermission(s.ctx, 151, SchemaPermission{Name: "TEST_PERM2", WithGrantOption: true})
}

func (s *SchemaTestSuite) TestUpdatePermissionRevokeGrantOption() {
	s.expectSchemaNameQuery("test_schema_update", int(s.schema.GetId(s.ctx)))
	s.dbMock.expectUsernameLookup(4567, "grant_user")
	expectExactExec(s.mock, "REVOKE GRANT OPTION FOR TEST_PERM ON schema::[test_schema_update] FROM [grant_user] CASCADE")

	s.schema.UpdatePermission(s.ctx, 4567, SchemaPermission{Name: "TEST_PERM", WithGrantOption: false})
}

func (s *SchemaTestSuite) TestRevokePermission() {
	s.expectSchemaNameQuery("test_schema_revoke", int(s.schema.GetId(s.ctx)))
	s.dbMock.expectUsernameLookup(96, "revoke_user")
	expectExactExec(s.mock, "REVOKE TEST_PERM4 ON schema::[test_schema_revoke] FROM [revoke_user] CASCADE")

	s.schema.RevokePermission(s.ctx, 96, "TEST_PERM4")
}

func (s *SchemaTestSuite) expectSchemaIdQuery(name string, id int) {
	expectExactQuery(s.mock, "SELECT SCHEMA_ID(@p1)").WithArgs(name).WillReturnRows(newRows("id").AddRow(id))
}

func (s *SchemaTestSuite) expectSchemaNameQuery(name string, id int) {
	expectExactQuery(s.mock, "SELECT SCHEMA_NAME(@p1)").WithArgs(id).WillReturnRows(newRows("name").AddRow(name))
}
