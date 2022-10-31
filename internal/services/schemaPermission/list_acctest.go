package schemaPermission

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testListDataSource(testCtx *acctest.TestContext) {
	testCtx.ExecDefaultDB("CREATE SCHEMA [test_perm_list]")
	defer testCtx.ExecDefaultDB("DROP SCHEMA [test_perm_list]")

	testCtx.ExecDefaultDB("CREATE ROLE [test_perm_list_role]")
	defer testCtx.ExecDefaultDB("DROP ROLE [test_perm_list_role]")

	testCtx.ExecDefaultDB("GRANT DELETE ON schema::[test_perm_list] TO [test_perm_list_role]")
	testCtx.ExecDefaultDB("GRANT ALTER ON schema::[test_perm_list] TO [test_perm_list_role] WITH GRANT OPTION")

	var schemaId, roleId int
	err := testCtx.GetDefaultDBConnection().
		QueryRow("SELECT SCHEMA_ID('test_perm_list'), DATABASE_PRINCIPAL_ID('test_perm_list_role')").
		Scan(&schemaId, &roleId)
	testCtx.Require.NoError(err, "Fetching IDs")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "mssql_schema_permissions" "test" {
	schema_id = %q
	principal_id = %q
}
`, testCtx.DefaultDbId(schemaId), testCtx.DefaultDbId(roleId)),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.mssql_schema_permissions.test", "permissions.*", map[string]string{
						"permission":        "DELETE",
						"with_grant_option": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.mssql_schema_permissions.test", "permissions.*", map[string]string{
						"permission":        "ALTER",
						"with_grant_option": "true",
					}),
				),
			},
		},
	})
}
