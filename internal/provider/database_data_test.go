package provider

import (
	sql2 "database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/PGSSoft/terraform-provider-mssql/internal/sql"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestDatabaseData(t *testing.T) {
	const resourceName = "data.mssql_database.test"
	var dbId string
	dbSettings := sql.DatabaseSettings{Name: "data_test_db", Collation: "SQL_Latin1_General_CP1_CS_AS"}

	newDataResource := func(name string) string {
		return fmt.Sprintf(`
data "mssql_database" "test" {
	name = %q
}`, name)
	}

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		PreCheck: func() {
			dbId = fmt.Sprint(createDB(t, dbSettings.Name))
			withDBConnection("master", func(conn *sql2.DB) {
				_, err := conn.Exec("ALTER DATABASE [data_test_db] COLLATE SQL_Latin1_General_CP1_CS_AS")
				require.NoError(t, err, "Setting DB collation")
			})
		},
		Steps: []resource.TestStep{
			{
				Config:      newDataResource("not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				Config: newDataResource("data_test_db"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceName, "id", &dbId),
					resource.TestCheckResourceAttr(resourceName, "name", dbSettings.Name),
					resource.TestCheckResourceAttr(resourceName, "collation", dbSettings.Collation),
				),
			},
		},
	})
}
