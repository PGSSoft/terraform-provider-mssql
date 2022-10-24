package schema

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
)

func testDataSource(testCtx *acctest.TestContext) {
	testCtx.ExecDefaultDB("CREATE ROLE test_schema_ds_owner")
	testCtx.ExecDefaultDB("CREATE SCHEMA test_schema_ds AUTHORIZATION test_schema_ds_owner")
	defer func() {
		testCtx.ExecDefaultDB("DROP SCHEMA test_schema_ds")
		testCtx.ExecDefaultDB("DROP ROLE test_schema_ds_owner")
	}()

	var schemaId, ownerId string
	err := testCtx.GetDefaultDBConnection().QueryRow("SELECT SCHEMA_ID('test_schema_ds'), USER_ID('test_schema_ds_owner')").Scan(&schemaId, &ownerId)
	testCtx.Require.NoError(err, "Fetching IDs")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "mssql_schema" "by_name" {
	database_id = %d
	name = "test_schema_ds"
}
`, testCtx.DefaultDBId),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_schema.by_name", "id", testCtx.DefaultDbId(schemaId)),
					resource.TestCheckResourceAttr("data.mssql_schema.by_name", "owner_id", testCtx.DefaultDbId(ownerId)),
				),
			},
			{
				Config: fmt.Sprintf(`
data "mssql_schema" "by_id" {
	id = %q
}
`, testCtx.DefaultDbId(schemaId)),

				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mssql_schema.by_id", "name", "test_schema_ds"),
					resource.TestCheckResourceAttr("data.mssql_schema.by_id", "database_id", fmt.Sprint(testCtx.DefaultDBId)),
					resource.TestCheckResourceAttr("data.mssql_schema.by_id", "owner_id", testCtx.DefaultDbId(ownerId)),
				),
			},
			{
				Config: fmt.Sprintf(`
data "mssql_schema" "not_exist" {
	database_id = %d
	name = "not_existsing_schema"
}
`, testCtx.DefaultDBId),

				ExpectError: regexp.MustCompile("not exist"),
			},
		},
	})
}
