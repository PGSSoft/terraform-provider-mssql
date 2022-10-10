package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestDatabaseRoleListData(t *testing.T) {
	var roleResourceId, ownerResourceId string

	defer testCtx.ExecDefaultDB(t, `
DROP ROLE [test_role];
DROP ROLE [test_owner];
		`)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testCtx.NewProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "mssql_database_roles" "master" {}`,
				Check: resource.TestCheckTypeSetElemNestedAttrs("data.mssql_database_roles.master", "roles.*", map[string]string{
					"id":          "1/0",
					"name":        "public",
					"database_id": "1",
					"owner_id":    "1/1",
				}),
			},
			{
				PreConfig: func() {
					conn := testCtx.GetDefaultDBConnection()
					var roleId, ownerId int
					err := conn.QueryRow(`
CREATE ROLE test_owner;
CREATE ROLE test_role AUTHORIZATION test_owner;
SELECT DATABASE_PRINCIPAL_ID('test_role'), DATABASE_PRINCIPAL_ID('test_owner');
`).Scan(&roleId, &ownerId)

					require.NoError(t, err, "creating role")

					roleResourceId = fmt.Sprintf("%d/%d", testCtx.DefaultDBId, roleId)
					ownerResourceId = fmt.Sprintf("%d/%d", testCtx.DefaultDBId, ownerId)
				},
				Config: fmt.Sprintf(`
data "mssql_database_roles" "test" {
	database_id = %d
}
`, testCtx.DefaultDBId),
				Check: func(state *terraform.State) error {
					return resource.TestCheckTypeSetElemNestedAttrs("data.mssql_database_roles.test", "roles.*", map[string]string{
						"id":          roleResourceId,
						"name":        "test_role",
						"database_id": fmt.Sprint(testCtx.DefaultDBId),
						"owner_id":    ownerResourceId,
					})(state)
				},
			},
		},
	})
}
