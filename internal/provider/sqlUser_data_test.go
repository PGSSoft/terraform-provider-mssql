package provider

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestSqlUserData(t *testing.T) {
	const dbName = "sql_user_data_test"
	var dbId, resourceId, userId, loginId string

	newDataResource := func(resourceName string, userName string) string {
		return fmt.Sprintf(`
data "mssql_database" %[1]q {
	name = %[2]q
}

data "mssql_sql_user" %[1]q {
	name = %[3]q
	database_id = data.mssql_database.%[1]s.id
}
`, resourceName, dbName, userName)
	}

	dataChecks := func(resName string) resource.TestCheckFunc {
		resName = fmt.Sprintf("data.mssql_sql_user.%s", resName)
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrPtr(resName, "id", &resourceId),
			resource.TestCheckResourceAttrPtr(resName, "login_id", &loginId),
			resource.TestCheckResourceAttrPtr(resName, "database_id", &dbId),
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		PreCheck: func() {
			dbId = fmt.Sprint(createDB(t, dbName))
		},
		Steps: []resource.TestStep{
			{
				Config:      newDataResource("not_exists", "not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				PreConfig: func() {
					withDBConnection("master", func(conn *sql.DB) {
						err := conn.QueryRow(`
CREATE LOGIN [test_login_sql_user_data] WITH PASSWORD='ComplictedPa$$w0rd13';
SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name]='test_login_sql_user_data'
`).Scan(&loginId)

						require.NoError(t, err, "creating login")
					})

					withDBConnection(dbName, func(conn *sql.DB) {
						err := conn.QueryRow(`
CREATE USER [test_user] FOR LOGIN [test_login_sql_user_data];
SELECT DATABASE_PRINCIPAL_ID('test_user')
`).Scan(&userId)

						require.NoError(t, err, "creating user")
					})

					resourceId = fmt.Sprintf("%s/%s", dbId, userId)
				},
				Config: newDataResource("exists", "test_user"),
				Check:  dataChecks("exists"),
			},
			{
				Config: `
data "mssql_sql_user" "master" {
	name = "dbo"
}
`,
				Check: resource.ComposeTestCheckFunc(
					sqlCheck("master", func(db *sql.DB) error {
						return db.QueryRow("SELECT principal_id, CONVERT(VARCHAR(85), [sid], 1) from sys.database_principals WHERE [name]='dbo'").
							Scan(&userId, &loginId)
					}),
					sqlCheck("master", func(db *sql.DB) error {
						err := db.QueryRow("SELECT database_id FROM sys.databases WHERE [name]='master'").Scan(&dbId)
						resourceId = fmt.Sprintf("%s/%s", dbId, userId)
						return err
					}),
					dataChecks("master"),
				),
			},
		},
	})
}
