package script

import (
	"database/sql"
	"fmt"
	"github.com/PGSSoft/terraform-provider-mssql/internal/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testResource(testCtx *acctest.TestContext) {
	columnResource := fmt.Sprintf(`
resource "mssql_script" "test_column" {
	database_id = %[1]d

	create_script = "ALTER TABLE test_table ADD test_column VARCHAR(MAX)"
	update_script = "ALTER TABLE test_table ALTER COLUMN test_column VARCHAR(MAX)"
	delete_script = "ALTER TABLE test_table DROP COLUMN test_column"

	read_script = <<SQL
SELECT 
	(SELECT COUNT(*) FROM sys.columns WHERE [object_id] = OBJECT_ID('test_table') AND [name] = 'test_column') AS [exists],
	(SELECT [system_type_id] FROM sys.columns WHERE [object_id] = OBJECT_ID('test_table') AND [name] = 'test_column') AS [type]
SQL

	state = {
		exists = "1"
		type = "167"
	}
}
`, testCtx.DefaultDBId)

	newTableConfig := func(resourceName string, tableName string) string {
		return fmt.Sprintf(`
resource "mssql_script" %[2]q {
	database_id = %[1]d
	read_script = "SELECT COUNT(*) AS [exists] FROM sys.tables WHERE [name]='%[3]s'"
	update_script = "CREATE TABLE %[3]s (id UNIQUEIDENTIFIER)"
	
	state = {
		exists = "1"
	}
}
`, testCtx.DefaultDBId, resourceName, tableName)
	}

	assertColumnType := testCtx.SqlCheckDefaultDB(func(conn *sql.DB) error {
		var typeId int
		err := conn.QueryRow("SELECT [system_type_id] FROM sys.columns WHERE [name] = 'test_column' AND [object_id] = OBJECT_ID('test_table')").Scan(&typeId)
		if err != nil {
			return err
		}

		testCtx.Assert.Equal(167, typeId, "Column type ID")

		return nil
	})

	testCtx.Test(resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: newTableConfig("test", "test_table"),
				Check: testCtx.SqlCheckDefaultDB(func(conn *sql.DB) error {
					return conn.QueryRow("SELECT [name] FROM sys.tables WHERE [name]=@p1", "test_table").Err()
				}),
			},
			{
				Config: columnResource,
				Check:  assertColumnType,
			},
			{
				PreConfig: func() {
					testCtx.ExecDefaultDB("ALTER TABLE [test_table] ALTER COLUMN [test_column] NVARCHAR(MAX)")
				},
				Config: columnResource,
				Check:  assertColumnType,
			},
			{
				Destroy: true,
				Config:  columnResource,
				Check: testCtx.SqlCheckDefaultDB(func(conn *sql.DB) error {
					var colCount int

					err := conn.QueryRow("SELECT COUNT(*) FROM sys.columns WHERE [name] = 'test_column' AND [object_id] = OBJECT_ID('test_table')").Scan(&colCount)
					if err != nil {
						return err
					}

					testCtx.Assert.Equal(0, colCount, "column count")

					return nil
				}),
			},
		},
	})
}
