package provider

import (
	"database/sql"
	"fmt"
	"testing"

	sql2 "github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSqlLoginResource(t *testing.T) {

	var newResourceDefaults = func(resourceName string, name string, password string) string {
		return fmt.Sprintf(`
resource "mssql_sql_login" %[1]q {
	name = %[2]q
	password = %[3]q
}
`, resourceName, name, password)
	}

	var newResource = func(resourceName string, settings sql2.SqlLoginSettings, defaultDbName string) string {
		loginAttributes := fmt.Sprintf(`
	must_change_password = %[1]v
	default_database_id = data.mssql_database.default.id
	default_language = %[2]q
	check_password_expiration = %[3]v
	check_password_policy = %[4]v
`,
			settings.MustChangePassword,
			settings.DefaultLanguage,
			settings.CheckPasswordExpiration,
			settings.CheckPasswordPolicy)

		if isAzureTest {
			loginAttributes = ""
		}

		return fmt.Sprintf(`
data "mssql_database" "default" {
	name = %[4]q
}

resource "mssql_sql_login" %[1]q {
	name = %[2]q
	password = %[3]q
	%[5]s
}
`,
			resourceName,
			settings.Name,
			settings.Password,
			defaultDbName,
			loginAttributes)
	}

	var loginId, defaultDbId, defaultLang string

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: newResourceDefaults("test_login", "login1", "Test_password123$"),
				Check: resource.ComposeTestCheckFunc(
					sqlCheck("master", func(db *sql.DB) error {
						return db.QueryRow(`	SELECT CONVERT(VARCHAR(85), [sid], 1), DB_ID(default_database_name), default_language_name 
													FROM sys.sql_logins 
													WHERE [name] = 'login1'`).
							Scan(&loginId, &defaultDbId, &defaultLang)
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_sql_login.test_login", "id", &loginId),
						resource.TestCheckResourceAttr("mssql_sql_login.test_login", "name", "login1"),
						resource.TestCheckResourceAttr("mssql_sql_login.test_login", "password", "Test_password123$"),
						sqlCheck("master", func(db *sql.DB) error {
							if isAzureTest {
								return nil
							}

							var mustChangePassword, checkPasswordExpiration, checkPasswordPolicy bool
							err := db.QueryRow(`	SELECT LOGINPROPERTY([name], 'IsMustChange'), is_expiration_checked, is_policy_checked 
														FROM sys.sql_logins 
														WHERE [name] = 'login1'`).
								Scan(&mustChangePassword, &checkPasswordExpiration, &checkPasswordPolicy)

							require.NoError(t, err, "SQL query")

							assert.False(t, mustChangePassword, "must_change_password")
							assert.False(t, checkPasswordExpiration, "check_password_expiration")
							assert.True(t, checkPasswordPolicy, "check_password_policy")

							return err
						}),
					),
				),
			},
			{
				PreConfig: func() {
					defaultDbId = fmt.Sprint(createDB(t, "test_db_login"))
				},
				Config: newResource("test_login_full", sql2.SqlLoginSettings{
					Name:                    "login2",
					Password:                "Str0ngPa$$w0rd124",
					CheckPasswordPolicy:     false,
					CheckPasswordExpiration: false,
					DefaultLanguage:         "polish",
					MustChangePassword:      false,
				}, "test_db_login"),
				Check: resource.ComposeTestCheckFunc(
					sqlCheck("master", func(db *sql.DB) error {
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
							if isAzureTest {
								return nil
							}

							return resource.ComposeAggregateTestCheckFunc(
								resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "must_change_password", "false"),
								resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "check_password_expiration", "false"),
								resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "check_password_policy", "false"),
								resource.TestCheckResourceAttrPtr("mssql_sql_login.test_login_full", "default_database_id", &defaultDbId),
								resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "default_language", "polish"),
								sqlCheck("master", func(db *sql.DB) error {
									var mustChangePassword, checkPasswordExpiration, checkPasswordPolicy bool
									var defaultDb string
									err := db.QueryRow(`	SELECT LOGINPROPERTY([name], 'IsMustChange'), is_expiration_checked, is_policy_checked, default_database_name, default_language_name
														FROM sys.sql_logins 
														WHERE [name] = 'login2'`).
										Scan(&mustChangePassword, &checkPasswordExpiration, &checkPasswordPolicy, &defaultDb, &defaultLang)

									require.NoError(t, err, "SQL query")

									assert.False(t, mustChangePassword, "must_change_password")
									assert.False(t, checkPasswordExpiration, "check_password_expiration")
									assert.False(t, checkPasswordPolicy, "check_password_policy")
									assert.Equal(t, "test_db_login", defaultDb, "default_database_id")
									assert.Equal(t, "polish", defaultLang)

									return err
								}),
							)(state)
						}),
				),
			},
			{
				PreConfig: func() {
					defaultDbId = fmt.Sprint(createDB(t, "test_db_login_2"))
				},
				Config: newResource("test_login_full", sql2.SqlLoginSettings{
					Name:                    "login3",
					Password:                "Test_password1234$",
					CheckPasswordPolicy:     true,
					CheckPasswordExpiration: true,
					DefaultLanguage:         "english",
					MustChangePassword:      true,
				}, "test_db_login_2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("mssql_sql_login.test_login_full", "id", &loginId),
					resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "name", "login3"),
					resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "password", "Test_password1234$"),
					func(state *terraform.State) error {
						if isAzureTest {
							return nil
						}

						return resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "must_change_password", "true"),
							resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "check_password_expiration", "true"),
							resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "check_password_policy", "true"),
							resource.TestCheckResourceAttrPtr("mssql_sql_login.test_login_full", "default_database_id", &defaultDbId),
							resource.TestCheckResourceAttr("mssql_sql_login.test_login_full", "default_language", "english"),
							sqlCheck("master", func(db *sql.DB) error {
								var mustChangePassword, checkPasswordExpiration, checkPasswordPolicy bool
								var defaultDb, name string

								err := db.QueryRow(`	SELECT [name], LOGINPROPERTY([name], 'IsMustChange'), is_expiration_checked, is_policy_checked, default_database_name, default_language_name
														FROM sys.sql_logins 
														WHERE CONVERT(VARCHAR(85), [sid], 1) = @p1`, loginId).
									Scan(&name, &mustChangePassword, &checkPasswordExpiration, &checkPasswordPolicy, &defaultDb, &defaultLang)

								require.NoError(t, err, "SQL query")

								assert.Equal(t, "login3", name, "name")
								assert.True(t, mustChangePassword, "must_change_password")
								assert.True(t, checkPasswordExpiration, "check_password_expiration")
								assert.True(t, checkPasswordPolicy, "check_password_policy")
								assert.Equal(t, "test_db_login_2", defaultDb, "default_database_id")
								assert.Equal(t, "english", defaultLang, "default_language")

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
