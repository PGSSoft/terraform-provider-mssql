package azureADUser

import (
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"regexp"
	"strings"
)

func testDataSource(testCtx *acctest.TestContext) {
	if !testCtx.IsAzureTest {
		return
	}

	configWithName := func(resourceName string, userName string) string {
		return fmt.Sprintf(`
data "mssql_azuread_user" %[1]q {
	name 		= %[2]q
	database_id = %[3]d
}
`, resourceName, userName, testCtx.DefaultDBId)
	}

	configWithObjectId := func(resourceName string, objectId string) string {
		return fmt.Sprintf(`
data "mssql_azuread_user" %[1]q {
	user_object_id 	= %[2]q
	database_id		= %[3]d
}
`, resourceName, objectId, testCtx.DefaultDBId)
	}

	var userResourceId string

	defer testCtx.ExecDefaultDB("DROP USER [%s]", testCtx.AzureADTestGroup.Name)

	testCtx.Test(resource.TestCase{
		PreCheck: func() {
			conn := testCtx.GetDefaultDBConnection()
			_, err := conn.Exec(`
				DECLARE @SQL NVARCHAR(MAX) = 'CREATE USER [' + @p1 + '] WITH SID=' + (SELECT CONVERT(VARCHAR(85), CONVERT(VARBINARY(85), CAST(@p2 AS UNIQUEIDENTIFIER), 1), 1)) + ', TYPE=E';
				EXEC(@SQL)`, testCtx.AzureADTestGroup.Name, testCtx.AzureADTestGroup.Id)
			testCtx.Require.NoError(err, "Creating AAD user")

			var userId int
			err = conn.QueryRow("SELECT principal_id FROM sys.database_principals WHERE [name]=@p1", testCtx.AzureADTestGroup.Name).Scan(&userId)
			testCtx.Require.NoError(err, "Fetching AAD user ID")

			userResourceId = fmt.Sprintf("%d/%d", testCtx.DefaultDBId, userId)
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
				Config: configWithName("existing_name", testCtx.AzureADTestGroup.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("data.mssql_azuread_user.existing_name", "id", &userResourceId),
					resource.TestCheckResourceAttr("data.mssql_azuread_user.existing_name", "user_object_id", strings.ToUpper(testCtx.AzureADTestGroup.Id)),
				),
			},
			{
				Config: configWithObjectId("existing_id", testCtx.AzureADTestGroup.Id),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("data.mssql_azuread_user.existing_id", "id", &userResourceId),
					resource.TestCheckResourceAttr("data.mssql_azuread_user.existing_id", "name", testCtx.AzureADTestGroup.Name),
				),
			},
		},
	})
}
