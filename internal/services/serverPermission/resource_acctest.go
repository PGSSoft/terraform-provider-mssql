package serverPermission

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testResource(testCtx *acctest.TestContext) {
	if testCtx.IsAzureTest {
		return
	}

	testCtx.ExecMasterDB("CREATE SERVER ROLE [test_server_permissions]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_server_permissions]")
	var roleId string
	err := testCtx.GetMasterDBConnection().
		QueryRow("SELECT [principal_id] FROM sys.server_principals WHERE [name] = 'test_server_permissions'").
		Scan(&roleId)
	testCtx.Require.NoError(err, "Fetching ID")

	newResource := func(resName string, permission string, withGrantOption bool) string {
		attrs := ""

		if withGrantOption {
			attrs = "with_grant_option = true"
		}

		return fmt.Sprintf(`
resource "mssql_server_permission" %[1]q {
	principal_id = %[2]q
	permission = %[3]q
	%[4]s
}
`, resName, roleId, permission, attrs)
	}

	checkPermissionState := func(permission string, state string) resource.TestCheckFunc {
		return testCtx.SqlCheckMaster(func(conn *sql.DB) error {
			return conn.QueryRow(`
SELECT 1 
FROM sys.server_permissions 
WHERE [class]=100 AND [grantee_principal_id]=@p1 AND [permission_name]=@p2 AND [state]=@p3`, roleId, permission, state).Err()
		})
	}

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newResource("test", "ALTER ANY DATABASE", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPermissionState("ALTER ANY DATABASE", "G"),
					resource.TestCheckResourceAttr("mssql_server_permission.test", "id", fmt.Sprintf("%s/ALTER ANY DATABASE", roleId)),
					resource.TestCheckResourceAttr("mssql_server_permission.test", "with_grant_option", "false"),
				),
			},
			{
				Config: newResource("test", "ALTER ANY DATABASE", true),
				Check:  checkPermissionState("ALTER ANY DATABASE", "W"),
			},
			{
				Config: newResource("test_with_grant", "ALTER ANY CONNECTION", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPermissionState("ALTER ANY DATABASE", "W"),
					resource.TestCheckResourceAttr("mssql_server_permission.test_with_grant", "id", fmt.Sprintf("%s/ALTER ANY CONNECTION", roleId)),
				),
			},
			{
				ResourceName:      "mssql_server_permission.test_with_grant",
				Config:            newResource("test_with_grant", "ALTER ANY CONNECTION", false),
				ImportStateId:     fmt.Sprintf("%s/ALTER ANY CONNECTION", roleId),
				ImportState:       true,
				ImportStateVerify: true,
				PlanOnly:          true,
			},
		},
	})
}
