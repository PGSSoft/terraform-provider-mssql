package serverRole

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testListDataSource(testCtx *acctest.TestContext) {
	testCtx.ExecMasterDB("CREATE SERVER ROLE [test_owner]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_owner]")

	testCtx.ExecMasterDB("CREATE SERVER ROLE [test_list_role] AUTHORIZATION [test_owner]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_list_role]")

	ownerId := fetchPrincipalId(testCtx, "'test_owner'")
	roleId := fetchPrincipalId(testCtx, "'test_list_role'")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `data "mssql_server_roles" "all" {}`,
				Check: resource.TestCheckTypeSetElemNestedAttrs("data.mssql_server_roles.all", "roles.*", map[string]string{
					"id":       roleId,
					"name":     "test_list_role",
					"owner_id": ownerId,
				}),
			},
		},
	})
}
