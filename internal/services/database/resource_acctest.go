package database

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testResource(testCtx *acctest.TestContext) {
	newDatabaseResource := func(resourceName string, dbName string) string {
		return fmt.Sprintf(`
resource "mssql_database" %[1]q {
	name = %[2]q
}
`, resourceName, dbName)
	}

	newDatabaseResourceWithCollation := func(resourceName string, dbName string, collation string) string {
		return fmt.Sprintf(`
	resource "mssql_database" %[1]q {
		name = %[2]q
		collation = %[3]q
	}
	`, resourceName, dbName, collation)
	}

	var dbId, dbCollation string

	var checkCollation = func(dbName string, expected string) resource.TestCheckFunc {
		return testCtx.SqlCheckMaster(func(db *sql.DB) error {
			var collation string
			err := db.QueryRow("SELECT collation_name FROM sys.databases WHERE name = @p1", dbName).Scan(&collation)
			testCtx.Assert.Equal(expected, collation)
			return err
		})
	}

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newDatabaseResource("test", "new_db"),
				Check: resource.ComposeTestCheckFunc(
					testCtx.SqlCheckMaster(func(db *sql.DB) error {
						return db.QueryRow("SELECT database_id, collation_name FROM sys.databases WHERE name = 'new_db'").Scan(&dbId, &dbCollation)
					}),
					resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrPtr("mssql_database.test", "id", &dbId),
						resource.TestCheckResourceAttr("mssql_database.test", "name", "new_db"),
						resource.TestCheckResourceAttrPtr("mssql_database.test", "collation", &dbCollation),
					),
				),
			},
			{
				Config: newDatabaseResourceWithCollation("new_db_with_collation", "new_db_with_collation", "SQL_Latin1_General_CP1_CS_AS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_database.new_db_with_collation", "collation", "SQL_Latin1_General_CP1_CS_AS"),
					checkCollation("new_db_with_collation", "SQL_Latin1_General_CP1_CS_AS"),
				),
			},
			{
				Config: newDatabaseResourceWithCollation("new_db_with_collation", "renamed_db_with_collation", "SQL_Latin1_General_CP1250_CI_AS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mssql_database.new_db_with_collation", "name", "renamed_db_with_collation"),
					resource.TestCheckResourceAttr("mssql_database.new_db_with_collation", "collation", "SQL_Latin1_General_CP1250_CI_AS"),
					checkCollation("renamed_db_with_collation", "SQL_Latin1_General_CP1250_CI_AS"),
				),
			},
			{
				ResourceName: "mssql_database.imported_db",
				Config:       newDatabaseResourceWithCollation("imported_db", "renamed_db_with_collation", "SQL_Latin1_General_CP1250_CI_AS"),
				ImportState:  true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					master := testCtx.GetMasterDBConnection()

					var id string
					err := master.QueryRow("SELECT database_id FROM sys.databases WHERE [name] = @p1", "renamed_db_with_collation").Scan(&id)
					return id, err
				},
				ImportStateVerify: true,
			},
		},
	})
}
