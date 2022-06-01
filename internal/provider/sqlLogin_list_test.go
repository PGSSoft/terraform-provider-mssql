package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestSqlLoginListData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "mssql_sql_logins" "list" {}`,
				Check: resource.TestCheckTypeSetElemNestedAttrs("data.mssql_sql_logins.list", "logins.*", map[string]string{
					"id":                        "0x01",
					"name":                      "sa",
					"must_change_password":      "false",
					"default_database_id":       "1",
					"default_language":          "us_english",
					"check_password_expiration": "false",
					"check_password_policy":     "true",
				}),
			},
		},
	})
}
