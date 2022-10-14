package azureADServicePrincipal

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
data "mssql_azuread_service_principal" %[1]q {
	name 		= %[2]q
	database_id = %[3]d
}
`, resourceName, userName, testCtx.DefaultDBId)
	}

	configWithObjectId := func(resourceName string, objectId string) string {
		return fmt.Sprintf(`
data "mssql_azuread_service_principal" %[1]q {
	client_id 	= %[2]q
	database_id	= %[3]d
}
`, resourceName, objectId, testCtx.DefaultDBId)
	}

	var userResourceId string

	defer testCtx.ExecDefaultDB("DROP USER [%s]", testCtx.AzureTestMSI.Name)

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					conn := testCtx.GetDefaultDBConnection()

					_, err := conn.Exec(`
DECLARE @SQL NVARCHAR(MAX) = 'CREATE USER [' + @p1 + '] WITH SID=' + (SELECT CONVERT(VARCHAR(85), CONVERT(VARBINARY(85), CAST(@p2 AS UNIQUEIDENTIFIER), 1), 1)) + ', TYPE=E';
EXEC(@SQL)`, testCtx.AzureTestMSI.Name, testCtx.AzureTestMSI.ClientId)
					testCtx.Require.NoError(err, "Creating AAD user")

					var userId int
					err = conn.QueryRow("SELECT principal_id FROM sys.database_principals WHERE [name]=@p1", testCtx.AzureTestMSI.Name).Scan(&userId)
					testCtx.Require.NoError(err, "Fetching AAD user ID")

					userResourceId = fmt.Sprintf("%d/%d", testCtx.DefaultDBId, userId)
				},
				Config:      configWithName("not_existing_name", "not_existing_name"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				Config:      configWithObjectId("not_existing_id", "a80e3c16-88a3-4218-ab27-4e25ef196bbf"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				Config: configWithName("existing_name", testCtx.AzureTestMSI.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("data.mssql_azuread_service_principal.existing_name", "id", &userResourceId),
					resource.TestCheckResourceAttr("data.mssql_azuread_service_principal.existing_name", "client_id", strings.ToUpper(testCtx.AzureTestMSI.ClientId)),
				),
			},
			{
				Config: configWithObjectId("existing_id", testCtx.AzureTestMSI.ClientId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr("data.mssql_azuread_service_principal.existing_id", "id", &userResourceId),
					resource.TestCheckResourceAttr("data.mssql_azuread_service_principal.existing_id", "name", testCtx.AzureTestMSI.Name),
				),
			},
		},
	})
}
