package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestDatabaseListData(t *testing.T) {
	const resourceName = "data.mssql_databases.list"

	var checkPredefinedDB = func(id int, name string) resource.TestCheckFunc {
		return resource.TestCheckTypeSetElemNestedAttrs(resourceName, "databases.*", map[string]string{
			"id":        fmt.Sprint(id),
			"name":      name,
			"collation": "SQL_Latin1_General_CP1_CI_AS",
		})
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
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
