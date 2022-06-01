package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestSqlLoginData(t *testing.T) {

	newDataResource := func(resourceName string, loginName string) string {
		return fmt.Sprintf(`
data "mssql_sql_login" %[1]q {
	name = %[2]q
}
`, resourceName, loginName)
	}

	var loginId, dbId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      newDataResource("not_exists", "not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				PreConfig: func() {
					db := openDBConnection()
					defer db.Close()

					_, err := db.Exec("CREATE DATABASE [test_db_login_data]")
					require.NoError(t, err, "creating DB")

					_, err = db.Exec("CREATE LOGIN [test_login] WITH PASSWORD='test_password', CHECK_POLICY=OFF, CHECK_EXPIRATION=OFF, DEFAULT_LANGUAGE=[polish], DEFAULT_DATABASE=[test_db_login_data]")
					require.NoError(t, err, "creating login")

					err = db.QueryRow("SELECT CONVERT(VARCHAR(85), SUSER_SID('test_login'), 1), DB_ID('test_db_login_data')").Scan(&loginId, &dbId)
					require.NoError(t, err, "fetching IDs")
				},
				Config: newDataResource("exists", "test_login"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("data.mssql_sql_login.exists", "id", &loginId),
					resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "name", "test_login"),
					resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "must_change_password", "false"),
					resource.TestCheckResourceAttrPtr("data.mssql_sql_login.exists", "default_database_id", &dbId),
					resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "default_language", "polish"),
					resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "check_password_expiration", "false"),
					resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "check_password_policy", "false"),
				),
			},
		},
	})
}
