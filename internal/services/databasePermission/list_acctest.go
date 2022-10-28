package databasePermission

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testListDataSource(testCtx *acctest.TestContext) {
	testCtx.ExecDefaultDB("CREATE ROLE [test_db_permissions_list]")
	defer testCtx.ExecDefaultDB("DROP ROLE [test_db_permissions_list]")

	var roleId int
	err := testCtx.GetDefaultDBConnection().QueryRow("SELECT principal_id FROM sys.database_principals WHERE [name]='test_db_permissions_list'").Scan(&roleId)
	testCtx.Require.NoError(err, "Fetching IDs")

	testCtx.ExecDefaultDB("GRANT DELETE TO [test_db_permissions_list]; GRANT ALTER TO [test_db_permissions_list] WITH GRANT OPTION")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "mssql_database_permissions" "test" {
	principal_id = %q
}
`, testCtx.DefaultDbId(roleId)),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.mssql_database_permissions.test", "permissions.*", map[string]string{
						"permission":        "DELETE",
						"with_grant_option": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.mssql_database_permissions.test", "permissions.*", map[string]string{
						"permission":        "ALTER",
						"with_grant_option": "true",
					}),
				),
			},
		},
	})
}
