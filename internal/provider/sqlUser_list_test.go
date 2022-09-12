package provider

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestSqlUserListData(t *testing.T) {
	var userId, loginId, resourceId string

	defer execMasterDB(t, "DROP LOGIN [sql_users_list_test]")
	defer execDefaultDB(t, "DROP USER [sql_users_list_test]")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
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
					withMasterDBConnection(func(conn *sql.DB) {
						err := conn.QueryRow(`
CREATE LOGIN [sql_users_list_test] WITH PASSWORD='C0mplicatedPa$$w0rd123';
SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name] = 'sql_users_list_test'
`).Scan(&loginId)

						require.NoError(t, err, "creating login")
					})

					withDefaultDBConnection(func(conn *sql.DB) {
						err := conn.QueryRow(`
CREATE USER [sql_users_list_test] FOR LOGIN [sql_users_list_test];
SELECT DATABASE_PRINCIPAL_ID('sql_users_list_test');
`).Scan(&userId)

						require.NoError(t, err, "creating user")
					})

					resourceId = fmt.Sprintf("%d/%s", defaultDbId, userId)
				},
				Config: fmt.Sprintf(`
data "mssql_database" "test" {
	name = %[1]q
}

data "mssql_sql_users" "test" {
	database_id = data.mssql_database.test.id
}
`, defaultDbName),
				Check: func(state *terraform.State) error {
					return resource.TestCheckTypeSetElemNestedAttrs("data.mssql_sql_users.test", "users.*", map[string]string{
						"id":          resourceId,
						"name":        "sql_users_list_test",
						"database_id": fmt.Sprint(defaultDbId),
						"login_id":    loginId,
					})(state)
				},
			},
		},
	})
}
