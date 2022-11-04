package sqlLogin

import (
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testListDataSource(testCtx *acctest.TestContext) {
	var loginId, principalId string

	defer testCtx.ExecMasterDB("DROP LOGIN [test_sql_login_list]")

	testCtx.Test(resource.TestCase{
		PreCheck: func() {
			conn := testCtx.GetMasterDBConnection()
			err := conn.QueryRow(`
CREATE LOGIN [test_sql_login_list] WITH PASSWORD='Str0ngPa$$w0rd124';
SELECT CONVERT(VARCHAR(85), [sid], 1), [principal_id] FROM sys.sql_logins WHERE [name]='test_sql_login_list'
`).Scan(&loginId, &principalId)

			testCtx.Require.NoError(err, "creating login")
		},
		Steps: []resource.TestStep{
			{
				Config: `data "mssql_sql_logins" "list" {}`,
				Check: func(state *terraform.State) error {
					expectedAttributes := map[string]string{
						"id":                        loginId,
						"name":                      "test_sql_login_list",
						"principal_id":              principalId,
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
