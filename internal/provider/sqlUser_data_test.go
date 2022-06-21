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

	var resourceId, userId, dbId, loginId string

	dataChecks := func(resName string) resource.TestCheckFunc {
		resName = fmt.Sprintf("data.mssql_sql_user.%s", resName)
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrPtr(resName, "id", &resourceId),
			resource.TestCheckResourceAttrPtr(resName, "login_id", &loginId),
			resource.TestCheckResourceAttrPtr(resName, "database_id", &dbId),
		)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					db := openDBConnection()
					defer db.Close()

					_, err := db.Exec(fmt.Sprintf("CREATE DATABASE [%s]", dbName))
					require.NoError(t, err, "creating DB")
				},
				Config:      newDataResource("not_exists", "not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				PreConfig: func() {
					db := openDBConnection()
					defer db.Close()

					err := db.QueryRow(fmt.Sprintf(`
USE [%s]; 
CREATE LOGIN [test_login_sql_user_data] WITH PASSWORD='test_password', CHECK_POLICY=OFF; 
CREATE USER [test_user] FOR LOGIN [test_login_sql_user_data];
SELECT DB_ID(), DATABASE_PRINCIPAL_ID('test_user'), CONVERT(VARCHAR(85), SUSER_SID('test_login_sql_user_data'), 1)`, dbName)).Scan(&dbId, &userId, &loginId)

					require.NoError(t, err, "creating user")

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
					sqlCheck(func(db *sql.DB) error {
						err := db.QueryRow("SELECT DB_ID(), DATABASE_PRINCIPAL_ID('dbo'), CONVERT(VARCHAR(85), SUSER_SID('sa'), 1)").
							Scan(&dbId, &userId, &loginId)

						resourceId = fmt.Sprintf("%s/%s", dbId, userId)

						return err
					}),
					dataChecks("master"),
				),
			},
		},
	})
}
