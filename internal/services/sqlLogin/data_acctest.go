package sqlLogin

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
)

func testDataSource(testCtx *acctest.TestContext) {

	newDataResource := func(resourceName string, loginName string) string {
		return fmt.Sprintf(`
data "mssql_sql_login" %[1]q {
	name = %[2]q
}
`, resourceName, loginName)
	}

	var loginId string

	defer testCtx.ExecMasterDB("DROP LOGIN [test_login]")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      newDataResource("not_exists", "not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				PreConfig: func() {
					conn := testCtx.GetMasterDBConnection()
					loginOptions := fmt.Sprintf(" MUST_CHANGE, CHECK_EXPIRATION=ON, CHECK_POLICY=ON, DEFAULT_LANGUAGE=[polish], DEFAULT_DATABASE=[%s]", acctest.DefaultDbName)
					if testCtx.IsAzureTest {
						loginOptions = ""
					}
					_, err := conn.Exec("CREATE LOGIN [test_login] WITH PASSWORD='C0mplicatedPa$$w0rd123'" + loginOptions)
					testCtx.Require.NoError(err, "creating login")

					err = conn.QueryRow("SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name] = 'test_login'").Scan(&loginId)
					testCtx.Require.NoError(err, "fetching IDs")
				},
				Config: newDataResource("exists", "test_login"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("data.mssql_sql_login.exists", "id", &loginId),
					resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "name", "test_login"),
					func(state *terraform.State) error {
						if testCtx.IsAzureTest {
							return nil
						}

						return resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "must_change_password", "true"),
							resource.TestCheckResourceAttr("data.mssql_sql_login.exists", "default_database_id", fmt.Sprint(testCtx.DefaultDBId)),
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
