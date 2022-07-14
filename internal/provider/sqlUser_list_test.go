package provider

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSqlUserListData(t *testing.T) {
	var dbId, userId, loginId, resourceId string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		PreCheck: func() {
			dbId = fmt.Sprint(createDB(t, "sql_users_list_test"))
		},
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
					withDBConnection("master", func(conn *sql.DB) {
						err := conn.QueryRow(`
CREATE LOGIN [sql_users_list_test] WITH PASSWORD='C0mplicatedPa$$w0rd123';
SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name] = 'sql_users_list_test'
`).Scan(&loginId)

						require.NoError(t, err, "creating login")
					})

					withDBConnection("sql_users_list_test", func(conn *sql.DB) {
						err := conn.QueryRow(`
CREATE USER [sql_users_list_test] FOR LOGIN [sql_users_list_test];
SELECT DATABASE_PRINCIPAL_ID('sql_users_list_test');
`).Scan(&userId)

						require.NoError(t, err, "creating user")
					})

					resourceId = fmt.Sprintf("%s/%s", dbId, userId)
				},
				Config: `
data "mssql_database" "test" {
	name = "sql_users_list_test"
}

data "mssql_sql_users" "test" {
	database_id = data.mssql_database.test.id
}
`,
				Check: func(state *terraform.State) error {
					return resource.TestCheckTypeSetElemNestedAttrs("data.mssql_sql_users.test", "users.*", map[string]string{
						"id":          resourceId,
						"name":        "sql_users_list_test",
						"database_id": dbId,
						"login_id":    loginId,
					})(state)
				},
			},
		},
	})
}
