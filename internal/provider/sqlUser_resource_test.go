package provider

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSqlUserResource(t *testing.T) {
	var dbId, userId, resourceId, loginId string

	var createLogin = func(loginName string) string {
		db := openDBConnection("master")
		defer db.Close()

		_, err := db.Exec(fmt.Sprintf("CREATE LOGIN [%s] WITH PASSWORD='Pa$$w0rd12'", loginName))
		require.NoError(t, err, "creating new login")

		var id string
		err = db.QueryRow("SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.sql_logins WHERE [name]=@p1", loginName).Scan(&id)
		require.NoError(t, err, "fetching login ID")

		return id
	}

	var newResource = func(resourceName string, name string, loginName string) string {
		return fmt.Sprintf(`
data "mssql_database" %[1]q {
	name = "user_test_db"
}

data "mssql_sql_login" %[1]q {
	name = %[3]q
}

resource "mssql_sql_user" %[1]q {
	name = %[2]q
	database_id = data.mssql_database.%[1]s.id
	login_id = data.mssql_sql_login.%[1]s.id
}
`, resourceName, name, loginName)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		PreCheck: func() {
			dbId = fmt.Sprint(createDB(t, "user_test_db"))
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					loginId = createLogin("sqluser_test_login")
				},
				Config: newResource("test_user", "test_user", "sqluser_test_login"),
				Check: resource.ComposeTestCheckFunc(
					sqlCheck("user_test_db", func(db *sql.DB) error {
						if err := db.QueryRow("SELECT USER_ID(@p1)", "test_user").Scan(&userId); err != nil {
							return err
						}

						resourceId = fmt.Sprintf("%s/%s", dbId, userId)

						return nil
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_sql_user.test_user", "id", &resourceId),
						resource.TestCheckResourceAttrPtr("mssql_sql_user.test_user", "database_id", &dbId),
						resource.TestCheckResourceAttrPtr("mssql_sql_user.test_user", "login_id", &loginId),
						resource.TestCheckResourceAttr("mssql_sql_user.test_user", "name", "test_user"),
						sqlCheck("user_test_db", func(db *sql.DB) error {
							var actualLoginId string
							err := db.QueryRow("SELECT CONVERT(VARCHAR(85), [sid], 1) FROM sys.database_principals WHERE principal_id=@p1", userId).Scan(&actualLoginId)

							assert.Equal(t, loginId, actualLoginId, "login ID")

							return err
						}),
					),
				),
			},
			{
				PreConfig: func() {
					loginId = createLogin("renamed_login")
				},
				Config: newResource("test_user", "renamed_user", "renamed_login"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("mssql_sql_user.test_user", "id", &resourceId),
					resource.TestCheckResourceAttrPtr("mssql_sql_user.test_user", "login_id", &loginId),
					resource.TestCheckResourceAttr("mssql_sql_user.test_user", "name", "renamed_user"),
					sqlCheck("user_test_db", func(db *sql.DB) error {
						var actualName, actualLoginId string
						err := db.QueryRow("SELECT [name], CONVERT(VARCHAR(85), [sid], 1) FROM sys.database_principals WHERE principal_id=@p1", userId).Scan(&actualName, &actualLoginId)

						assert.Equal(t, "renamed_user", actualName)
						assert.Equal(t, loginId, actualLoginId)

						return err
					}),
				),
			},
			{
				ResourceName: "mssql_sql_user.test_user",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return resourceId, nil
				},
				ImportStateVerify: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					return nil
				},
			},
		},
	})
}
