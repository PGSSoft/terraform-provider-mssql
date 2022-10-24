package schema

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testListDataSource(testCtx *acctest.TestContext) {
	testCtx.ExecDefaultDB("CREATE SCHEMA [test_list]")
	defer testCtx.ExecDefaultDB("DROP SCHEMA [test_list]")

	var schemaId, ownerId int
	err := testCtx.GetDefaultDBConnection().QueryRow("SELECT schema_id, principal_id FROM sys.schemas WHERE [name]='test_list'").Scan(&schemaId, &ownerId)
	testCtx.Require.NoError(err, "Fetching IDs")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "mssql_schemas" "all" {
	database_id = %d
}
`, testCtx.DefaultDBId),

				Check: resource.TestCheckTypeSetElemNestedAttrs("data.mssql_schemas.all", "schemas.*", map[string]string{
					"id":          testCtx.DefaultDbId(schemaId),
					"owner_id":    testCtx.DefaultDbId(ownerId),
					"database_id": fmt.Sprint(testCtx.DefaultDBId),
					"name":        "test_list",
				}),
			},
		},
	})
}
