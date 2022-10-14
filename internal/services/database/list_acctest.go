package database

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testListDataSource(testCtx *acctest.TestContext) {
	const resourceName = "data.mssql_databases.list"

	var checkPredefinedDB = func(id int, name string) resource.TestCheckFunc {
		return resource.TestCheckTypeSetElemNestedAttrs(resourceName, "databases.*", map[string]string{
			"id":        fmt.Sprint(id),
			"name":      name,
			"collation": "SQL_Latin1_General_CP1_CI_AS",
		})
	}

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `data "mssql_databases" "list" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPredefinedDB(1, "master"),
				),
			},
		},
	})
}
