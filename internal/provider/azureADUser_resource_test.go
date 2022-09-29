package provider

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAADUserResource(t *testing.T) {
	if !isAzureTest {
		return
	}

	var userId int
	var userResourceId string

	newResource := func(resourceName string, name string) string {
		return fmt.Sprintf(`
data "mssql_database" %[1]q {
	name = %[4]q
}

resource "mssql_azuread_user" %[1]q {
	name = %[2]q
	database_id = data.mssql_database.%[1]s.id
	user_object_id = %[3]q
}
`, resourceName, name, azureAdTestGroupId, defaultDbName)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: newResource("test_user", "test_aad_user"),
				Check: resource.ComposeTestCheckFunc(
					sqlCheck(defaultDbName, func(db *sql.DB) error {
						if err := db.QueryRow("SELECT principal_id FROM sys.database_principals WHERE [name] = 'test_aad_user'").Scan(&userId); err != nil {
							return err
						}

						userResourceId = fmt.Sprintf("%d/%d", defaultDbId, userId)

						return nil
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_azuread_user.test_user", "id", &userResourceId),
						sqlCheck(defaultDbName, func(db *sql.DB) error {
							var userType, userSid string
							err := db.QueryRow("SELECT [type], CONVERT(VARCHAR(36), CONVERT(UNIQUEIDENTIFIER, [sid], 1), 1) FROM sys.database_principals WHERE principal_id = @p1", userId).
								Scan(&userType, &userSid)

							assert.Equal(t, "E", strings.ToUpper(userType), "user type")
							assert.Equal(t, strings.ToUpper(azureAdTestGroupId), strings.ToUpper(userSid), "user SID")

							return err
						}),
					),
				),
			},
			{
				ResourceName: "mssql_azuread_user.test_user",
				ImportState:  true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return userResourceId, nil
				},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					for _, state := range states {
						if state.ID == userResourceId {
							assert.Equal(t, "test_aad_user", state.Attributes["name"])
							assert.Equal(t, fmt.Sprint(defaultDbId), state.Attributes["database_id"])
							assert.Equal(t, strings.ToUpper(azureAdTestGroupId), strings.ToUpper(state.Attributes["user_object_id"]))
						}
					}

					return nil
				},
			},
		},
	})
}
