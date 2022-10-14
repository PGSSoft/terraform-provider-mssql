package sqlUser

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
)

func testDataSource(testCtx *acctest.TestContext) {
	var resourceId, userId, loginId string
	dbId := fmt.Sprint(testCtx.DefaultDBId)

	newDataResource := func(resourceName string, userName string) string {
		return fmt.Sprintf(`
data "mssql_sql_user" %[1]q {
	name = %[3]q
	database_id = %[2]d
}
`, resourceName, testCtx.DefaultDBId, userName)
	}

	dataChecks := func(resName string) resource.TestCheckFunc {
		resName = fmt.Sprintf("data.mssql_sql_user.%s", resName)
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrPtr(resName, "id", &resourceId),
			resource.TestCheckResourceAttrPtr(resName, "login_id", &loginId),
			resource.TestCheckResourceAttrPtr(resName, "database_id", &dbId),
		)
	}

	defer testCtx.ExecMasterDB("DROP LOGIN [test_login_sql_user_data]")
	defer testCtx.ExecDefaultDB("DROP USER [test_user]")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      newDataResource("not_exists", "not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				PreConfig: func() {
					master := testCtx.GetMasterDBConnection()
					err := master.QueryRow(`
CREATE LOGIN [test_login_sql_user_data] WITH PASSWORD='ComplictedPa$$w0rd13';
SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name]='test_login_sql_user_data'
`).Scan(&loginId)

					testCtx.Require.NoError(err, "creating login")

					defaultDB := testCtx.GetDefaultDBConnection()
					err = defaultDB.QueryRow(`
CREATE USER [test_user] FOR LOGIN [test_login_sql_user_data];
SELECT DATABASE_PRINCIPAL_ID('test_user')
`).Scan(&userId)

					testCtx.Require.NoError(err, "creating user")

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
					testCtx.SqlCheckMaster(func(db *sql.DB) error {
						return db.QueryRow("SELECT principal_id, CONVERT(VARCHAR(85), [sid], 1) from sys.database_principals WHERE [name]='dbo'").
							Scan(&userId, &loginId)
					}),
					testCtx.SqlCheckMaster(func(db *sql.DB) error {
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
