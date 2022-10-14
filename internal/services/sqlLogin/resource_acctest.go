package sqlLogin

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	sql2 "github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testResource(testCtx *acctest.TestContext) {

	var newResourceDefaults = func(resourceName string, name string, password string) string {
		return fmt.Sprintf(`
resource "mssql_sql_login" %[1]q {
	name = %[2]q
	password = %[3]q
}
`, resourceName, name, password)
	}

	var newResource = func(resourceName string, settings sql2.SqlLoginSettings, dbId int) string {
		loginAttributes := fmt.Sprintf(`
	must_change_password = %[1]v
	default_database_id = %[5]d
	default_language = %[2]q
	check_password_expiration = %[3]v
	check_password_policy = %[4]v
`,
			settings.MustChangePassword,
			settings.DefaultLanguage,
			settings.CheckPasswordExpiration,
			settings.CheckPasswordPolicy,
			dbId)

		if testCtx.IsAzureTest {
			loginAttributes = ""
		}

		return fmt.Sprintf(`
resource "mssql_sql_login" %[1]q {
	name = %[2]q
	password = %[3]q
	%[4]s
}
`,
			resourceName,
			settings.Name,
			settings.Password,
			loginAttributes)
	}

	var loginId, defaultLang, loginDefaultDbId string

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newResourceDefaults("test_login", "login1", "Test_password123$"),
				Check: resource.ComposeTestCheckFunc(
					testCtx.SqlCheckMaster(func(db *sql.DB) error {
						return db.QueryRow(`	SELECT CONVERT(VARCHAR(85), [sid], 1), DB_ID(default_database_name), default_language_name 
													FROM sys.sql_logins 
													WHERE [name] = 'login1'`).
							Scan(&loginId, &loginDefaultDbId, &defaultLang)
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_sql_login.test_login", "id", &loginId),
						resource.TestCheckResourceAttr("mssql_sql_login.test_login", "name", "login1"),
						resource.TestCheckResourceAttr("mssql_sql_login.test_login", "password", "Test_password123$"),
						testCtx.SqlCheckMaster(func(db *sql.DB) error {
							if testCtx.IsAzureTest {
								return nil
							}

							var mustChangePassword, checkPasswordExpiration, checkPasswordPolicy bool
							err := db.QueryRow(`	SELECT LOGINPROPERTY([name], 'IsMustChange'), is_expiration_checked, is_policy_checked 
														FROM sys.sql_logins 
														WHERE [name] = 'login1'`).
								Scan(&mustChangePassword, &checkPasswordExpiration, &checkPasswordPolicy)

							testCtx.Require.NoError(err, "SQL query")

							testCtx.Assert.False(mustChangePassword, "must_change_password")
							testCtx.Assert.False(checkPasswordExpiration, "check_password_expiration")
							testCtx.Assert.True(checkPasswordPolicy, "check_password_policy")

							return err
						}),
					),
				),
			},
			{
				Config: newResource("test_login_full", sql2.SqlLoginSettings{
					Name:                    "login2",
					Password:                "Str0ngPa$$w0rd124",
					CheckPasswordPolicy:     false,
					CheckPasswordExpiration: false,
					DefaultLanguage:         "polish",
					MustChangePassword:      false,
				}, testCtx.DefaultDBId),
				Check: resource.ComposeTestCheckFunc(
					testCtx.SqlCheckMaster(func(db *sql.DB) error {
						return db.QueryRow(`	SELECT CONVERT(VARCHAR(85), [sid], 1) 
													FROM sys.sql_logins 
													WHERE [name] = 'login2'`).
							Scan(&loginId)
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_sql_login.test_login_full", "id", &loginId),
						resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "name", "login2"),
						resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "password", "Str0ngPa$$w0rd124"),
						func(state *terraform.State) error {
							if testCtx.IsAzureTest {
								return nil
							}

							return resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "must_change_password", "false"),
								resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "check_password_expiration", "false"),
								resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "check_password_policy", "false"),
								resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "default_database_id", fmt.Sprint(testCtx.DefaultDBId)),
								resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "default_language", "polish"),
								testCtx.SqlCheckMaster(func(db *sql.DB) error {
									var mustChangePassword, checkPasswordExpiration, checkPasswordPolicy bool
									var defaultDb string
									err := db.QueryRow(`	SELECT LOGINPROPERTY([name], 'IsMustChange'), is_expiration_checked, is_policy_checked, default_database_name, default_language_name
														FROM sys.sql_logins 
														WHERE [name] = 'login2'`).
										Scan(&mustChangePassword, &checkPasswordExpiration, &checkPasswordPolicy, &defaultDb, &defaultLang)

									testCtx.Require.NoError(err, "SQL query")

									testCtx.Assert.False(mustChangePassword, "must_change_password")
									testCtx.Assert.False(checkPasswordExpiration, "check_password_expiration")
									testCtx.Assert.False(checkPasswordPolicy, "check_password_policy")
									testCtx.Assert.Equal(acctest.DefaultDbName, defaultDb, "default_database_id")
									testCtx.Assert.Equal("polish", defaultLang)

									return err
								}),
							)(state)
						}),
				),
			},
			{
				Config: newResource("test_login_full", sql2.SqlLoginSettings{
					Name:                    "login3",
					Password:                "Test_password1234$",
					CheckPasswordPolicy:     true,
					CheckPasswordExpiration: true,
					DefaultLanguage:         "english",
					MustChangePassword:      true,
				}, testCtx.DefaultDBId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("mssql_sql_login.test_login_full", "id", &loginId),
					resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "name", "login3"),
					resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "password", "Test_password1234$"),
					func(state *terraform.State) error {
						if testCtx.IsAzureTest {
							return nil
						}

						return resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "must_change_password", "true"),
							resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "check_password_expiration", "true"),
							resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "check_password_policy", "true"),
							resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "default_database_id", fmt.Sprint(testCtx.DefaultDBId)),
							resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "default_language", "english"),
							testCtx.SqlCheckMaster(func(db *sql.DB) error {
								var mustChangePassword, checkPasswordExpiration, checkPasswordPolicy bool
								var defaultDb, name string

								err := db.QueryRow(`	SELECT [name], LOGINPROPERTY([name], 'IsMustChange'), is_expiration_checked, is_policy_checked, default_database_name, default_language_name
														FROM sys.sql_logins 
														WHERE CONVERT(VARCHAR(85), [sid], 1) = @p1`, loginId).
									Scan(&name, &mustChangePassword, &checkPasswordExpiration, &checkPasswordPolicy, &defaultDb, &defaultLang)

								testCtx.Require.NoError(err, "SQL query")

								testCtx.Assert.Equal("login3", name, "name")
								testCtx.Assert.True(mustChangePassword, "must_change_password")
								testCtx.Assert.True(checkPasswordExpiration, "check_password_expiration")
								testCtx.Assert.True(checkPasswordPolicy, "check_password_policy")
								testCtx.Assert.Equal(acctest.DefaultDbName, defaultDb, "default_database_id")
								testCtx.Assert.Equal("english", defaultLang, "default_language")

								return err
							}),
						)(state)
					},
				),
			},
			{
				ResourceName: "mssql_sql_login.imported",
				Config:       newResourceDefaults("imported", "login3", "Test_password1234$"),
				ImportState:  true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return loginId, nil
				},
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
					"check_password_expiration",
					"check_password_policy",
					"default_database_id",
					"default_language",
					"must_change_password",
				},
			},
		},
	})
}
