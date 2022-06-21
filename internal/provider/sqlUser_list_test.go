package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSqlUserListData(t *testing.T) {
	var dbId, userId, loginId, resourceId string

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
					db := openDBConnection()
					defer db.Close()

					_, err := db.Exec("CREATE DATABASE [sql_users_list_test]")
					require.NoError(t, err, "creating DB")

					err = db.QueryRow(`
USE [sql_users_list_test];
CREATE LOGIN [sql_users_list_test] WITH PASSWORD='test_password', CHECK_POLICY=OFF;
CREATE USER [sql_users_list_test] FOR LOGIN [sql_users_list_test];
SELECT DB_ID(), DATABASE_PRINCIPAL_ID('sql_users_list_test'), CONVERT(VARCHAR(85), SUSER_SID('sql_users_list_test'), 1);
`).Scan(&dbId, &userId, &loginId)

					require.NoError(t, err, "creating user")

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
