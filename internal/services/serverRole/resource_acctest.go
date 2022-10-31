package serverRole

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testResource(testCtx *acctest.TestContext) {
	newResource := func(resName string, name string, ownerId string) string {
		attrs := ""
		if ownerId != "" {
			attrs += fmt.Sprintf("owner_id = %q", ownerId)
		}

		return fmt.Sprintf(`
resource "mssql_server_role" %[1]q {
	name = %[2]q
	%[3]s
}
`, resName, name, attrs)
	}

	testCtx.ExecMasterDB("CREATE SERVER ROLE [test_role_owner]")
	defer testCtx.ExecMasterDB("DROP SERVER ROLE [test_role_owner]")

	testOwnerId := fetchPrincipalId(testCtx, "'test_role_owner'")
	currentLoginId := fetchPrincipalId(testCtx, "ORIGINAL_LOGIN()")
	var actualRoleId, actualOwnerId string

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newResource("test", "test_role", ""),
				Check: resource.ComposeTestCheckFunc(
					testCtx.SqlCheckMaster(func(conn *sql.DB) error {
						err := conn.
							QueryRow("SELECT [principal_id], [owning_principal_id] FROM sys.server_principals WHERE [name]='test_role'").
							Scan(&actualRoleId, &actualOwnerId)

						testCtx.Assert.Equal(currentLoginId, actualOwnerId, "owner")

						return err
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_server_role.test", "id", &actualRoleId),
						resource.TestCheckResourceAttr("mssql_server_role.test", "owner_id", currentLoginId),
					),
				),
			},
			{
				Config: newResource("test", "renamed_role", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCtx.SqlCheckMaster(func(conn *sql.DB) error {
						var name string
						err := conn.
							QueryRow("SELECT [name] FROM sys.server_principals WHERE [principal_id]=@p1", actualRoleId).
							Scan(&name)

						testCtx.Assert.Equal("renamed_role", name, "name")

						return err
					}),
					resource.TestCheckResourceAttrPtr("mssql_server_role.test", "id", &actualRoleId),
				),
			},
			{
				Config: newResource("with_owner", "test_role_with_owner", testOwnerId),
				Check: resource.ComposeTestCheckFunc(
					testCtx.SqlCheckMaster(func(conn *sql.DB) error {
						err := conn.
							QueryRow("SELECT [principal_id], [owning_principal_id] FROM sys.server_principals WHERE [name]='test_role_with_owner'").
							Scan(&actualRoleId, &actualOwnerId)

						testCtx.Assert.Equal(testOwnerId, actualOwnerId, "owner")

						return err
					}),
					resource.TestCheckResourceAttrPtr("mssql_server_role.with_owner", "id", &actualRoleId),
				),
			},
			{
				ResourceName:       "mssql_server_role.with_owner",
				Config:             newResource("with_owner", "test_role_with_owner", testOwnerId),
				PlanOnly:           true,
				ImportState:        true,
				ImportStateVerify:  true,
				ImportStatePersist: false,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return actualRoleId, nil
				},
			},
		},
	})
}
