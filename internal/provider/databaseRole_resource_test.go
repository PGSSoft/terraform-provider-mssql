package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseRoleResource(t *testing.T) {
	var roleId, roleResourceId, ownerResourceId string

	newResource := func(resourceName string, roleName string, ownerRoleName string) string {
		return fmt.Sprintf(`
data "mssql_database" %[1]q {
	name = %[4]q
}

resource "mssql_database_role" %[3]q {
	name = %[3]q
	database_id = data.mssql_database.%[1]s.id

	lifecycle {
		create_before_destroy = true
	}
}

resource "mssql_database_role" %[1]q {
	name = %[2]q
	database_id = data.mssql_database.%[1]s.id
	owner_id = mssql_database_role.%[3]s.id
}
`, resourceName, roleName, ownerRoleName, defaultDbName)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: newResource("with_owner", "test_with_owner", "test_owner"),
				Check: resource.ComposeTestCheckFunc(
					sqlCheck(defaultDbName, func(db *sql.DB) error {
						var ownerName, userId, userName string

						err := db.QueryRow("SELECT DATABASE_PRINCIPAL_ID('test_with_owner'), DATABASE_PRINCIPAL_ID()").Scan(&roleId, &userId)
						if err != nil {
							return err
						}

						roleResourceId = fmt.Sprintf("%d/%s", defaultDbId, roleId)
						ownerResourceId = fmt.Sprintf("%d/%s", defaultDbId, userId)

						err = db.QueryRow("SELECT USER_NAME(owning_principal_id) FROM sys.database_principals WHERE [name] = 'test_with_owner'").Scan(&ownerName)
						if err != nil {
							return err
						}

						assert.Equal(t, "test_owner", ownerName, "explicit owner")

						err = db.QueryRow("SELECT USER_NAME(owning_principal_id), USER_NAME() FROM sys.database_principals WHERE [name] = 'test_owner'").Scan(&ownerName, &userName)
						if err != nil {
							return err
						}

						assert.Equal(t, userName, ownerName, "implicit owner")

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
					sqlCheck(defaultDbName, func(db *sql.DB) error {
						var name, ownerName string
						err := db.QueryRow("SELECT [name], USER_NAME(owning_principal_id) FROM sys.database_principals WHERE principal_id = @p1", roleId).
							Scan(&name, &ownerName)
						if err != nil {
							return err
						}

						assert.Equal(t, "renamed", name, "name")
						assert.Equal(t, "new_owner", ownerName, "owner")

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
							assert.Equal(t, "renamed", state.Attributes["name"])
						}
					}
					return nil
				},
			},
		},
	})
}
