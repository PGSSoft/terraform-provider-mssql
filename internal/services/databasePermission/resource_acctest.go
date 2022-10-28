package databasePermission

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testResource(testCtx *acctest.TestContext) {
	testCtx.ExecDefaultDB("CREATE ROLE [test_db_permissions]")
	defer testCtx.ExecDefaultDB("DROP ROLE [test_db_permissions]")

	var roleId int
	err := testCtx.GetDefaultDBConnection().QueryRow("SELECT principal_id FROM sys.database_principals WHERE [name]='test_db_permissions'").Scan(&roleId)
	testCtx.Require.NoError(err, "Fetching IDs")

	checkPermissionState := func(permission string, expectedState string) resource.TestCheckFunc {
		return testCtx.SqlCheckDefaultDB(func(conn *sql.DB) error {
			var state string
			err := conn.QueryRow("SELECT [state] FROM sys.database_permissions WHERE [grantee_principal_id]=@p1 AND [permission_name]=@p2", roleId, permission).Scan(&state)

			testCtx.Assert.Equal(expectedState, state, "permission state")

			return err
		})
	}

	newResource := func(resName string, permission string, withGrantOption bool) string {
		additionalAttrs := ""

		if withGrantOption {
			additionalAttrs = "with_grant_option = true"
		}

		return fmt.Sprintf(`
resource "mssql_database_permission" %[1]q {
	principal_id = %[2]q
	permission = %[3]q
	%[4]s
}
`, resName, testCtx.DefaultDbId(roleId), permission, additionalAttrs)
	}

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newResource("test", "CREATE TABLE", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPermissionState("CREATE TABLE", "G"),
					resource.TestCheckResourceAttr("mssql_database_permission.test", "with_grant_option", "false"),
					resource.TestCheckResourceAttr("mssql_database_permission.test", "id", testCtx.DefaultDbId(roleId, "CREATE TABLE")),
				),
			},
			{
				Config: newResource("test", "CREATE TABLE", true),
				Check:  checkPermissionState("CREATE TABLE", "W"),
			},
			{
				Config: newResource("with_grant", "DELETE", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPermissionState("DELETE", "W"),
					resource.TestCheckResourceAttr("mssql_database_permission.with_grant", "id", testCtx.DefaultDbId(roleId, "DELETE")),
				),
			},
			{
				ResourceName:      "mssql_database_permission.with_grant",
				Config:            newResource("with_grant", "DELETE", true),
				ImportStateId:     testCtx.DefaultDbId(roleId, "DELETE"),
				ImportState:       true,
				ImportStateVerify: true,
				PlanOnly:          true,
			},
		},
	})
}
