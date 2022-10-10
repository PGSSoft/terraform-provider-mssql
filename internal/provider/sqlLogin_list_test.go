package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestSqlLoginListData(t *testing.T) {
	var loginId string

	defer testCtx.ExecMasterDB(t, "DROP LOGIN [test_sql_login_list]")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testCtx.NewProviderFactories(),
		PreCheck: func() {
			conn := testCtx.GetMasterDBConnection()
			err := conn.QueryRow(`
CREATE LOGIN [test_sql_login_list] WITH PASSWORD='Str0ngPa$$w0rd124';
SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name]='test_sql_login_list'
`).Scan(&loginId)

			require.NoError(t, err, "creating login")
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

					if testCtx.IsAzureTest {
						expectedAttributes["check_password_policy"] = "false"
					}

					return resource.TestCheckTypeSetElemNestedAttrs("data.mssql_sql_logins.list", "logins.*", expectedAttributes)(state)
				},
			},
		},
	})
}
