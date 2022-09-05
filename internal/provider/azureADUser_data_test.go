package provider

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAzureADUserData(t *testing.T) {
	if !isAzureTest {
		return
	}

	configWithName := func(resourceName string, userName string) string {
		return fmt.Sprintf(`
data "mssql_database" %[1]q {
	name = "aad_user_data"
}

data "mssql_azuread_user" %[1]q {
	name 		= %[2]q
	database_id = data.mssql_database.%[1]s.id
}
`, resourceName, userName)
	}

	configWithObjectId := func(resourceName string, objectId string) string {
		return fmt.Sprintf(`
data "mssql_database" %[1]q {
	name = "aad_user_data"
}

data "mssql_azuread_user" %[1]q {
	user_object_id 	= %[2]q
	database_id		= data.mssql_database.%[1]s.id
}
`, resourceName, objectId)
	}

	var userResourceId string
	var dbId int

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		PreCheck: func() {
			dbId = createDB(t, "aad_user_data")
			withDBConnection("aad_user_data", func(conn *sql.DB) {
				_, err := conn.Exec(`
				DECLARE @SQL NVARCHAR(MAX) = 'CREATE USER [' + @p1 + '] WITH SID=' + (SELECT CONVERT(VARCHAR(85), CONVERT(VARBINARY(85), CAST(@p2 AS UNIQUEIDENTIFIER), 1), 1)) + ', TYPE=E';
				EXEC(@SQL)`, azureAdTestGroupId)
				require.NoError(t, err, "Creating AAD user")

				var userId int
				err = conn.QueryRow("SELECT principal_id FROM sys.database_principals WHERE [name]=@p1", azureAdTestGroupName).Scan(&userId)
				require.NoError(t, err, "Fetching AAD user ID")

				userResourceId = fmt.Sprintf("%d/%d", dbId, userId)
			})
		},
		Steps: []resource.TestStep{
			{
				Config:      configWithName("not_existing_name", "not_existing_name"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				Config:      configWithObjectId("not_existing_id", "a80e3c16-88a3-4218-ab27-4e25ef196bbf"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				Config: configWithName("existing_name", azureMSIName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("data.mssql_azuread_user.existing_name", "id", &userResourceId),
					resource.TestCheckResourceAttr("data.mssql_azuread_user.existing_name", "user_object_id", strings.ToUpper(azureAdTestGroupId)),
				),
			},
			{
				Config: configWithObjectId("existing_id", azureAdTestGroupId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("data.mssql_azuread_user.existing_id", "id", &userResourceId),
					resource.TestCheckResourceAttr("data.mssql_azuread_user.existing_id", "name", azureMSIName),
				),
			},
		},
	})
}
