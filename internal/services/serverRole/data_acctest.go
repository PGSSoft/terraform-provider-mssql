package serverRole

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testDataSource(testCtx *acctest.TestContext) {
	testCtx.ExecMasterDB("CREATE SERVER ROLE [test_owner]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_owner]")

	testCtx.ExecMasterDB("CREATE SERVER ROLE [test_role_data] AUTHORIZATION [test_owner]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_role_data]")

	roleId := fetchPrincipalId(testCtx, "'test_role_data'")
	ownerId := fetchPrincipalId(testCtx, "'test_owner'")

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
				),
			},
		},
	})
}
