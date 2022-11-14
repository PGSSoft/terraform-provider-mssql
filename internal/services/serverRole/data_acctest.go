package serverRole

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testDataSource(testCtx *acctest.TestContext) {
	if testCtx.IsAzureTest {
		return
	}

	testCtx.ExecMasterDB("CREATE SERVER ROLE [test_owner]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_owner]")

	testCtx.ExecMasterDB("CREATE SERVER ROLE [test_role_data] AUTHORIZATION [test_owner]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_role_data]")

	testCtx.ExecMasterDB("CREATE SERVER ROLE [test_role_member]")
	testCtx.ExecMasterDB("ALTER SERVER ROLE [test_role_data] ADD MEMBER [test_role_member]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_role_member]")

	roleId := fetchPrincipalId(testCtx, "'test_role_data'")
	ownerId := fetchPrincipalId(testCtx, "'test_owner'")
	memberId := fetchPrincipalId(testCtx, "'test_role_member'")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
data "mssql_server_role" "by_name" {
	name = "test_role_data"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_server_role.by_name", "id", roleId),
					resource.TestCheckResourceAttr("data.mssql_server_role.by_name", "owner_id", ownerId),
					resource.TestCheckTypeSetElemNestedAttrs("data.mssql_server_role.by_name", "members.*", map[string]string{
						"id":   memberId,
						"name": "test_role_member",
						"type": "SERVER_ROLE",
					}),
				),
			},
			{
				Config: fmt.Sprintf(`
data "mssql_server_role" "by_id" {
	id = %q
}
`, roleId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_server_role.by_id", "name", "test_role_data"),
					resource.TestCheckResourceAttr("data.mssql_server_role.by_id", "owner_id", ownerId),
					resource.TestCheckTypeSetElemNestedAttrs("data.mssql_server_role.by_id", "members.*", map[string]string{
						"id":   memberId,
						"name": "test_role_member",
						"type": "SERVER_ROLE",
					}),
				),
			},
		},
	})
}
