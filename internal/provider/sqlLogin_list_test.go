package provider

import (
	"database/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSqlLoginListData(t *testing.T) {
	var loginId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		PreCheck: func() {
			withDBConnection("master", func(conn *sql.DB) {
				err := conn.QueryRow(`
CREATE LOGIN [test_sql_login_list] WITH PASSWORD='Str0ngPa$$w0rd124';
SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name]='test_sql_login_list'
`).Scan(&loginId)

				require.NoError(t, err, "creating login")
			})
		},
		Steps: []resource.TestStep{
			{
				Config: `data "mssql_sql_logins" "list" {}`,
				Check: func(state *terraform.State) error {
					expectedAttributes := map[string]string{
						"id":                        loginId,
						"name":                      "test_sql_login_list",
						"must_change_password":      "false",
						"default_database_id":       "1",
						"default_language":          "us_english",
						"check_password_expiration": "false",
						"check_password_policy":     "true",
					}

					if isAzureTest {
						expectedAttributes["check_password_policy"] = "false"
					}

					return resource.TestCheckTypeSetElemNestedAttrs("data.mssql_sql_logins.list", "logins.*", expectedAttributes)(state)
				},
			},
		},
	})
}
