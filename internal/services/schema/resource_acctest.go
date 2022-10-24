package schema

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testResource(testCtx *acctest.TestContext) {
	var schemaId, ownerId string
	var testRoleId, dboId int

	newResource := func(resName string, schemaName string, ownerId string) string {
		attrs := ""

		if ownerId != "" {
			attrs = fmt.Sprintf("owner_id=%q", ownerId)
		}

		return fmt.Sprintf(`
resource "mssql_schema" %[1]q {
	database_id = %[3]d
	name = %[2]q
	%[4]s
}
`, resName, schemaName, testCtx.DefaultDBId, attrs)
	}

	testCtx.ExecDefaultDB("CREATE ROLE test_schema_owner")
	err := testCtx.GetDefaultDBConnection().
		QueryRow("SELECT principal_id FROM sys.database_principals WHERE [name] = 'test_schema_owner'").
		Scan(&testRoleId)
	testCtx.Require.NoError(err, "Fetching owner ID")

	defer testCtx.ExecDefaultDB("DROP ROLE test_schema_owner")

	err = testCtx.GetDefaultDBConnection().QueryRow("SELECT principal_id FROM sys.database_principals WHERE [name] = 'dbo'").Scan(&dboId)
	testCtx.Require.NoError(err, "Fetching dbo ID")

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newResource("default_owner", "default_owner", ""),
				Check: resource.ComposeTestCheckFunc(
					testCtx.SqlCheckDefaultDB(func(conn *sql.DB) error {
						var userId, sId int
						err := conn.QueryRow("SELECT USER_ID(), SCHEMA_ID('default_owner')").Scan(&userId, &sId)

						schemaId = testCtx.DefaultDbId(sId)
						ownerId = testCtx.DefaultDbId(userId)

						return err
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_schema.default_owner", "id", &schemaId),
						resource.TestCheckResourceAttrPtr("mssql_schema.default_owner", "owner_id", &ownerId),
					),
				),
			},
			{
				Config: newResource("with_owner", "with_owner", testCtx.DefaultDbId(testRoleId)),
				Check: resource.ComposeTestCheckFunc(
					testCtx.SqlCheckDefaultDB(func(conn *sql.DB) error {
						var sId, pId int
						err := conn.QueryRow("SELECT schema_id, principal_id FROM sys.schemas WHERE [name] = 'with_owner'").Scan(&sId, &pId)
						schemaId = testCtx.DefaultDbId(sId)

						testCtx.Assert.Equal(testRoleId, pId, "Schema owner")

						return err
					}),
					resource.TestCheckResourceAttrPtr("mssql_schema.with_owner", "id", &schemaId),
				),
			},
			{
				Config: newResource("with_owner", "with_owner", testCtx.DefaultDbId(dboId)),
				Check: testCtx.SqlCheckDefaultDB(func(conn *sql.DB) error {
					var sId, pId int
					err := conn.QueryRow("SELECT schema_id, principal_id FROM sys.schemas WHERE [name] = 'with_owner'").Scan(&sId, &pId)

					testCtx.Assert.Equal(dboId, pId, "owner id")
					testCtx.Assert.Equal(schemaId, testCtx.DefaultDbId(sId), "schema id")

					return err
				}),
			},
			{
				ResourceName:       "mssql_schema.imported",
				Config:             newResource("imported", "with_owner", testCtx.DefaultDbId(dboId)),
				PlanOnly:           true,
				ImportState:        true,
				ImportStatePersist: false,
				ImportStateVerify:  true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					var sId int
					err := testCtx.GetDefaultDBConnection().QueryRow("SELECT SCHEMA_ID('with_owner')").Scan(&sId)

					return testCtx.DefaultDbId(sId), err
				},
			},
		},
	})
}
