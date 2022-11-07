package serverRoleMember

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testResource(testCtx *acctest.TestContext) {
	roleName := "###MS_ServerStateReader###"

	var memberId string
	err := testCtx.GetMasterDBConnection().QueryRow("SELECT [principal_id] FROM sys.server_principals WHERE [name] = ORIGINAL_LOGIN()").Scan(&memberId)
	testCtx.Require.NoError(err, "Fetching IDs")

	if !testCtx.IsAzureTest {
		testCtx.ExecMasterDB("CREATE SERVER ROLE [test_role_member]")
		defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_role_member]")
		roleName = "test_role_member"

		testCtx.ExecMasterDB("CREATE SERVER ROLE [test_role_member_member]")
		defer testCtx.ExecDefaultDB("DROP SERVER ROLE [test_role_member_member]")
		err = testCtx.GetMasterDBConnection().QueryRow("SELECT [principal_id] FROM sys.server_principals WHERE [name] = 'test_role_member_member'").Scan(&memberId)
		testCtx.Require.NoError(err, "Fetching IDs")
	}

	var roleId string
	err = testCtx.GetMasterDBConnection().QueryRow("SELECT [principal_id] FROM sys.server_principals WHERE [name]=@p1", roleName).Scan(&roleId)
	testCtx.Require.NoError(err, "Fetching IDs")

	resourceId := fmt.Sprintf("%s/%s", roleId, memberId)

	config := fmt.Sprintf(`
resource "mssql_server_role_member" "test" {
	role_id = %[1]q
	member_id = %[2]q
}
`, roleId, memberId)

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCtx.SqlCheckMaster(func(conn *sql.DB) error {
						return conn.QueryRow("SELECT 1 FROM sys.server_role_members WHERE [role_principal_id] = @p1 AND [member_principal_id] = @p2", roleId, memberId).Err()
					}),
					resource.TestCheckResourceAttr("mssql_server_role_member.test", "id", resourceId),
				),
			},
			{
				ImportState:        true,
				ImportStatePersist: false,
				ImportStateVerify:  true,
				ImportStateId:      resourceId,
				ResourceName:       "mssql_server_role_member.test",
				Config:             config,
				PlanOnly:           true,
			},
		},
	})
}
