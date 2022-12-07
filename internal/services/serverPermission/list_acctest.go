package serverPermission

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testListDataSource(testCtx *acctest.TestContext) {
	if testCtx.IsAzureTest {
		return
	}

	testCtx.ExecMasterDB("CREATE SERVER ROLE [server_perm_test]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [server_perm_test]")
	testCtx.ExecMasterDB("GRANT VIEW ANY DATABASE TO [server_perm_test]")
	testCtx.ExecMasterDB("GRANT VIEW SERVER STATE TO [server_perm_test] WITH GRANT OPTION")

	var principalId string
	err := testCtx.GetMasterDBConnection().QueryRow("SELECT [principal_id] FROM sys.server_principals WHERE [name]='server_perm_test'").Scan(&principalId)
	testCtx.Require.NoError(err, "Fetching principal ID")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "mssql_server_permissions" "test" {
	principal_id = %q
}
`, principalId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.mssql_server_permissions.test", "permissions.*", map[string]string{
						"permission":        "VIEW ANY DATABASE",
						"with_grant_option": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.mssql_server_permissions.test", "permissions.*", map[string]string{
						"permission":        "VIEW SERVER STATE",
						"with_grant_option": "true",
					}),
				),
			},
		},
	})
}
