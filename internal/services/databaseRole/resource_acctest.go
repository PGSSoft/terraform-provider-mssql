package databaseRole

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testResource(testCtx *acctest.TestContext) {
	var roleId, roleResourceId, ownerResourceId string

	newResource := func(resourceName string, roleName string, ownerRoleName string) string {
		return fmt.Sprintf(`
resource "mssql_database_role" %[3]q {
	name = %[3]q
	database_id = %[4]d

	lifecycle {
		create_before_destroy = true
	}
}

resource "mssql_database_role" %[1]q {
	name = %[2]q
	database_id = %[4]d
	owner_id = mssql_database_role.%[3]s.id
}
`, resourceName, roleName, ownerRoleName, testCtx.DefaultDBId)
	}

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newResource("with_owner", "test_with_owner", "test_owner"),
				Check: resource.ComposeTestCheckFunc(
					testCtx.SqlCheckDefaultDB(func(db *sql.DB) error {
						var ownerName, userId, userName string

						err := db.QueryRow("SELECT DATABASE_PRINCIPAL_ID('test_with_owner'), DATABASE_PRINCIPAL_ID()").Scan(&roleId, &userId)
						if err != nil {
							return err
						}

						roleResourceId = fmt.Sprintf("%d/%s", testCtx.DefaultDBId, roleId)
						ownerResourceId = fmt.Sprintf("%d/%s", testCtx.DefaultDBId, userId)

						err = db.QueryRow("SELECT USER_NAME(owning_principal_id) FROM sys.database_principals WHERE [name] = 'test_with_owner'").Scan(&ownerName)
						if err != nil {
							return err
						}

						testCtx.Assert.Equal("test_owner", ownerName, "explicit owner")

						err = db.QueryRow("SELECT USER_NAME(owning_principal_id), USER_NAME() FROM sys.database_principals WHERE [name] = 'test_owner'").Scan(&ownerName, &userName)
						if err != nil {
							return err
						}

						testCtx.Assert.Equal(userName, ownerName, "implicit owner")

						return nil
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_database_role.test_owner", "owner_id", &ownerResourceId),
						resource.TestCheckResourceAttrPtr("mssql_database_role.with_owner", "id", &roleResourceId),
					),
				),
			},
			{
				Config: newResource("with_owner", "renamed", "new_owner"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCtx.SqlCheckDefaultDB(func(db *sql.DB) error {
						var name, ownerName string
						err := db.QueryRow("SELECT [name], USER_NAME(owning_principal_id) FROM sys.database_principals WHERE principal_id = @p1", roleId).
							Scan(&name, &ownerName)
						if err != nil {
							return err
						}

						testCtx.Assert.Equal("renamed", name, "name")
						testCtx.Assert.Equal("new_owner", ownerName, "owner")

						return nil
					}),
					resource.TestCheckResourceAttrPtr("mssql_database_role.with_owner", "id", &roleResourceId),
				),
			},
			{
				ResourceName: "mssql_database_role.with_owner",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return roleResourceId, nil
				},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					for _, state := range states {
						if state.ID == roleResourceId {
							testCtx.Assert.Equal("renamed", state.Attributes["name"])
						}
					}
					return nil
				},
			},
		},
	})
}
