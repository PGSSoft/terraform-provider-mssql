package provider

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
		PreCheck: func() {
			dbId = fmt.Sprint(createDB(t, "test_db_login_data"))
		},
		Steps: []resource.TestStep{
			{
				Config:      newDataResource("not_exists", "not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				PreConfig: func() {
					withDBConnection("master", func(conn *sql.DB) {
						loginOptions := " MUST_CHANGE, CHECK_EXPIRATION=ON, CHECK_POLICY=ON, DEFAULT_LANGUAGE=[polish], DEFAULT_DATABASE=[test_db_login_data]"
						if isAzureTest {
							loginOptions = ""
						}
						_, err := conn.Exec("CREATE LOGIN [test_login] WITH PASSWORD='C0mplicatedPa$$w0rd123'" + loginOptions)
						require.NoError(t, err, "creating login")

						err = conn.QueryRow("SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name] = 'test_login'").Scan(&loginId)
						require.NoError(t, err, "fetching IDs")
					})
				},
				Config: newDataResource("exists", "test_login"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("data.mssql_sql_login.exists", "id", &loginId),
					resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "name", "test_login"),
					func(state *terraform.State) error {
						if isAzureTest {
							return nil
						}

						return resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "must_change_password", "true"),
							resource.TestCheckResourceAttrPtr("data.mssql_sql_login.exists", "default_database_id", &dbId),
							resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "default_language", "polish"),
							resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "check_password_expiration", "true"),
							resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "check_password_policy", "true"),
						)(state)
					},
				),
			},
		},
	})
}
