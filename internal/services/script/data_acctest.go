package script

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testDataSource(testCtx *acctest.TestContext) {

	newConfig := func(resName string, query string) string {
		return fmt.Sprintf(`
data "mssql_query" %[1]q {
	database_id = %[3]d
	query = %[2]q
}
`, resName, query, testCtx.DefaultDBId)
	}

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newConfig("multirow", "SELECT 1 AS X, 'FOO' AS Y UNION ALL SELECT 3, 'BAR'"),
				Check: func(state *terraform.State) error {
					testAttr := func(name string, value string) resource.TestCheckFunc {
						return resource.TestCheckResourceAttr("data.mssql_query.multirow", name, value)
					}

					return resource.ComposeAggregateTestCheckFunc(
						testAttr("result.#", "2"),
						testAttr("result.0.X", "1"),
						testAttr("result.0.Y", "FOO"),
						testAttr("result.1.X", "3"),
						testAttr("result.1.Y", "BAR"),
					)(state)
				},
			},
			{
				Config: newConfig("null", "SELECT 1 AS X, NULL AS Y"),
				Check:  resource.TestCheckNoResourceAttr("data.mssql_query.null", "result.0.Y"),
			},
			{
				Config: newConfig("no_rows", "SELECT 1 WHERE 1=0"),
				Check:  resource.TestCheckResourceAttr("data.mssql_query.no_rows", "result.#", "0"),
			},
		},
	})
}
