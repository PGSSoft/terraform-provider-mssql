package azureADServicePrincipal

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
)

func testResource(testCtx *acctest.TestContext) {
	if !testCtx.IsAzureTest {
		return
	}

	var userId int
	var userResourceId string

	newResource := func(resourceName string, name string) string {
		return fmt.Sprintf(`
resource "mssql_azuread_service_principal" %[1]q {
	name = %[2]q
	database_id = %[4]d
	client_id = %[3]q
}
`, resourceName, name, testCtx.AzureTestMSI.ClientId, testCtx.DefaultDBId)
	}

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newResource("test_user", "test_aad_user"),
				Check: resource.ComposeTestCheckFunc(
					testCtx.SqlCheckDefaultDB(func(db *sql.DB) error {
						if err := db.QueryRow("SELECT principal_id FROM sys.database_principals WHERE [name] = 'test_aad_user'").Scan(&userId); err != nil {
							return err
						}

						userResourceId = fmt.Sprintf("%d/%d", testCtx.DefaultDBId, userId)

						return nil
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_azuread_service_principal.test_user", "id", &userResourceId),
						testCtx.SqlCheckDefaultDB(func(db *sql.DB) error {
							var userType, userSid string
							err := db.QueryRow("SELECT [type], CONVERT(VARCHAR(36), CONVERT(UNIQUEIDENTIFIER, [sid], 1), 1) FROM sys.database_principals WHERE principal_id = @p1", userId).
								Scan(&userType, &userSid)

							testCtx.Assert.Equal("E", strings.ToUpper(userType), "user type")
							testCtx.Assert.Equal(strings.ToUpper(testCtx.AzureTestMSI.ClientId), strings.ToUpper(userSid), "user SID")

							return err
						}),
					),
				),
			},
			{
				ResourceName: "mssql_azuread_service_principal.test_user",
				ImportState:  true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return userResourceId, nil
				},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					for _, state := range states {
						if state.ID == userResourceId {
							testCtx.Assert.Equal("test_aad_user", state.Attributes["name"])
							testCtx.Assert.Equal(fmt.Sprint(testCtx.DefaultDBId), state.Attributes["database_id"])
							testCtx.Assert.Equal(strings.ToUpper(testCtx.AzureTestMSI.ClientId), strings.ToUpper(state.Attributes["client_id"]))
						}
					}

					return nil
				},
			},
		},
	})
}
