package schemaPermission

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testResource(testCtx *acctest.TestContext) {
	testCtx.ExecDefaultDB("CREATE SCHEMA [test_permission]")
	defer testCtx.ExecDefaultDB("DROP SCHEMA [test_permission]")

	testCtx.ExecDefaultDB("CREATE ROLE [test_schema_permission]")
	defer testCtx.ExecDefaultDB("DROP ROLE [test_schema_permission]")

	var schemaId, roleId int
	err := testCtx.GetDefaultDBConnection().
		QueryRow("SELECT SCHEMA_ID('test_permission'), DATABASE_PRINCIPAL_ID('test_schema_permission')").
		Scan(&schemaId, &roleId)
	testCtx.Require.NoError(err, "Fetching IDs")

	newResource := func(resName string, permission string, withGrantOption bool) string {
		additionalAttrs := ""

		if withGrantOption {
			additionalAttrs = "with_grant_option = true"
		}

		return fmt.Sprintf(`
resource "mssql_schema_permission" %[1]q {
	schema_id = %[2]q
	principal_id = %[3]q
	permission = %[4]q
	%[5]s
}
`, resName, testCtx.DefaultDbId(schemaId), testCtx.DefaultDbId(roleId), permission, additionalAttrs)
	}

	checkPermissionState := func(permission string, expectedState string) resource.TestCheckFunc {
		return testCtx.SqlCheckDefaultDB(func(conn *sql.DB) error {
			var state string
			err := conn.
				QueryRow("SELECT [state] FROM sys.database_permissions WHERE [class]=3 AND [major_id]=@p1 AND [grantee_principal_id]=@p2", schemaId, roleId).
				Scan(&state)

			testCtx.Assert.Equal(expectedState, state, "permission state")

			return err
		})
	}

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newResource("test", "ALTER", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPermissionState("ALTER", "G"),
					resource.TestCheckResourceAttr("mssql_schema_permission.test", "id", fmt.Sprintf("%d/%d/%d/ALTER", testCtx.DefaultDBId, schemaId, roleId)),
					resource.TestCheckResourceAttr("mssql_schema_permission.test", "with_grant_option", "false"),
				),
			},
			{
				Config: newResource("test", "ALTER", true),
				Check:  checkPermissionState("ALTER", "W"),
			},
			{
				Config: newResource("with_grant", "DELETE", true),
				Check:  checkPermissionState("DELETE", "W"),
			},
			{
				ResourceName:      "mssql_schema_permission.with_grant",
				Config:            newResource("with_grant", "DELETE", true),
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%d/%d/%d/DELETE", testCtx.DefaultDBId, schemaId, roleId),
				ImportStateVerify: true,
				PlanOnly:          true,
			},
		},
	})
}
