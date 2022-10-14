package databaseRole

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testListDataSource(testCtx *acctest.TestContext) {
	var roleResourceId, ownerResourceId string

	defer testCtx.ExecDefaultDB(`
DROP ROLE [test_role];
DROP ROLE [test_owner];
		`)

	testCtx.Test(resource.TestCase{
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

					testCtx.Require.NoError(err, "creating role")

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
