package provider

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestDatabaseRoleData(t *testing.T) {
	var roleResourceId, ownerResourceId, roleMemberResourceId, userMemberResourceId string
	var dbId string

	newDataResource := func(resourceName string, roleName string) string {
		return fmt.Sprintf(`
data "mssql_database" %[1]q {
	name = "db_role_data_test"
}

data "mssql_database_role" %[1]q {
	name = %[2]q
	database_id = data.mssql_database.%[1]s.id
}
`, resourceName, roleName)
	}

	formatId := func(id int) string { return fmt.Sprintf("%s/%d", dbId, id) }

	setIds := func(roleId int, ownerId int) {
		roleResourceId = formatId(roleId)
		ownerResourceId = formatId(ownerId)
	}

	attributesCheck := func(resourceName string) resource.TestCheckFunc {
		resourceName = fmt.Sprintf("data.mssql_database_role.%s", resourceName)
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrPtr(resourceName, "id", &roleResourceId),
			resource.TestCheckResourceAttrPtr(resourceName, "owner_id", &ownerResourceId),
			resource.TestCheckResourceAttrPtr(resourceName, "database_id", &dbId),
		)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		PreCheck: func() {
			dbId = fmt.Sprint(createDB(t, "db_role_data_test"))
		},
		Steps: []resource.TestStep{
			{
				Config:      newDataResource("not_exists", "not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				PreConfig: func() {
					withDBConnection("db_role_data_test", func(conn *sql.DB) {
						var roleId, ownerId, roleMemberId, userMemberId int

						err := conn.QueryRow(`
CREATE ROLE [test_owner];
CREATE ROLE [test_role_member];
CREATE USER [test_user_member] WITHOUT LOGIN;
CREATE ROLE [test_role] AUTHORIZATION [test_owner];
ALTER ROLE [test_role] ADD MEMBER [test_role_member];
ALTER ROLE [test_role] ADD MEMBER [test_user_member];
SELECT 
    DATABASE_PRINCIPAL_ID('test_role'), 
    DATABASE_PRINCIPAL_ID('test_owner'), 
    DATABASE_PRINCIPAL_ID('test_role_member'),
    DATABASE_PRINCIPAL_ID('test_user_member');
`).Scan(&roleId, &ownerId, &roleMemberId, &userMemberId)

						require.NoError(t, err, "creating role")
						setIds(roleId, ownerId)
						roleMemberResourceId = formatId(roleMemberId)
						userMemberResourceId = formatId(userMemberId)
					})
				},
				Config: newDataResource("exists", "test_role"),
				Check: resource.ComposeAggregateTestCheckFunc(
					attributesCheck("exists"),
					func(state *terraform.State) error {
						memberCheck := func(attrs map[string]string) resource.TestCheckFunc {
							return resource.TestCheckTypeSetElemNestedAttrs("data.mssql_database_role.exists", "members.*", attrs)
						}
						return resource.ComposeAggregateTestCheckFunc(
							memberCheck(map[string]string{
								"id":   roleMemberResourceId,
								"name": "test_role_member",
								"type": "DATABASE_ROLE",
							}),
							memberCheck(map[string]string{
								"id":   userMemberResourceId,
								"name": "test_user_member",
								"type": "SQL_USER",
							}),
						)(state)
					},
				),
			},
			{
				Config: `
data "mssql_database_role" "master" {
	name = "public"
}
`,
				Check: resource.ComposeTestCheckFunc(
					sqlCheck("master", func(db *sql.DB) error {
						var roleId, ownerId int
						err := db.QueryRow("SELECT database_id, DATABASE_PRINCIPAL_ID('public'), DATABASE_PRINCIPAL_ID('dbo') FROM sys.databases WHERE [name]='master'").
							Scan(&dbId, &roleId, &ownerId)
						setIds(roleId, ownerId)
						return err
					}),
					attributesCheck("master"),
				),
			},
		},
	})
}
