package provider

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestDatabaseRoleMemberResource(t *testing.T) {
	var resourceId string
	var dbId int

	newResource := func(resourceName string, roleName string, memberName string) string {
		return fmt.Sprintf(`
data "mssql_database" %[3]q {
	name = "db_role_member_test"
}

resource "mssql_database_role" %[3]q {
	name = %[3]q
	database_id = data.mssql_database.%[3]s.id
}

resource "mssql_database_role" %[1]q {
	name = %[2]q
	database_id = data.mssql_database.%[3]s.id
}

resource "mssql_database_role_member" %[1]q {
	role_id = mssql_database_role.%[1]s.id
	member_id = mssql_database_role.%[3]s.id
}
`, resourceName, roleName, memberName)
	}

	checkMembership := func(roleName string, memberName string) resource.TestCheckFunc {
		return resource.ComposeTestCheckFunc(
			sqlCheck("db_role_member_test", func(db *sql.DB) error {
				var roleId, memberId int

				err := db.QueryRow("SELECT DATABASE_PRINCIPAL_ID(@p1), DATABASE_PRINCIPAL_ID(@p2)", roleName, memberName).
					Scan(&roleId, &memberId)
				if err != nil {
					return err
				}

				resourceId = fmt.Sprintf("%d/%d/%d", dbId, roleId, memberId)

				return db.QueryRow("SELECT 1 FROM sys.database_role_members WHERE role_principal_id = @p1 AND member_principal_id = @p2", roleId, memberId).Err()
			}),
			resource.TestCheckResourceAttrPtr("mssql_database_role_member.new_resource", "id", &resourceId),
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		PreCheck: func() {
			dbId = createDB(t, "db_role_member_test")
		},
		Steps: []resource.TestStep{
			{
				Config: newResource("new_resource", "test_role", "test_member"),
				Check:  checkMembership("test_role", "test_member"),
			},
			{
				Config: newResource("new_resource", "test_role", "another_member"),
				Check:  checkMembership("test_role", "another_member"),
			},
			{
				ResourceName: "mssql_database_role_member.new_resource",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return resourceId, nil
				},
				ImportStateVerify: true,
			},
			{
				Config: `
data "mssql_database" "invalid" {
	name = "db_role_member_test"
}

data "mssql_database_role" "public" {
	name = "db_owner"
}

resource "mssql_database_role" "invalid_membership" {
	name = "invalid_membership"
	database_id = data.mssql_database.invalid.id
}

resource "mssql_database_role_member" "invalid" {
	member_id = mssql_database_role.invalid_membership.id
	role_id = data.mssql_database_role.public.id
}
`,
				ExpectError: regexp.MustCompile("same database"),
			},
		},
	})
}
