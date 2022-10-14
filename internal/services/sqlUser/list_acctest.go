package sqlUser

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testListDataSource(testCtx *acctest.TestContext) {
	var userId, loginId, resourceId string

	defer testCtx.ExecMasterDB("DROP LOGIN [sql_users_list_test]")
	defer testCtx.ExecDefaultDB("DROP USER [sql_users_list_test]")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `data "mssql_sql_users" "master" {}`,
				Check: resource.TestCheckTypeSetElemNestedAttrs("data.mssql_sql_users.master", "users.*", map[string]string{
					"id":          "1/1",
					"name":        "dbo",
					"database_id": "1",
					"login_id":    "0x01",
				}),
			},
			{
				PreConfig: func() {
					master := testCtx.GetMasterDBConnection()
					err := master.QueryRow(`
CREATE LOGIN [sql_users_list_test] WITH PASSWORD='C0mplicatedPa$$w0rd123';
SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name] = 'sql_users_list_test'
`).Scan(&loginId)

					testCtx.Require.NoError(err, "creating login")

					defaultDB := testCtx.GetDefaultDBConnection()
					err = defaultDB.QueryRow(`
CREATE USER [sql_users_list_test] FOR LOGIN [sql_users_list_test];
SELECT DATABASE_PRINCIPAL_ID('sql_users_list_test');
`).Scan(&userId)

					testCtx.Require.NoError(err, "creating user")

					resourceId = fmt.Sprintf("%d/%s", testCtx.DefaultDBId, userId)
				},
				Config: fmt.Sprintf(`
data "mssql_sql_users" "test" {
	database_id = %[1]d
}
`, testCtx.DefaultDBId),
				Check: func(state *terraform.State) error {
					return resource.TestCheckTypeSetElemNestedAttrs("data.mssql_sql_users.test", "users.*", map[string]string{
						"id":          resourceId,
						"name":        "sql_users_list_test",
						"database_id": fmt.Sprint(testCtx.DefaultDBId),
						"login_id":    loginId,
					})(state)
				},
			},
		},
	})
}
