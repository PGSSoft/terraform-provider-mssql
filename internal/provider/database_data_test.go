package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/tkielar/terraform-provider-mssql/internal/sql"
	"regexp"
	"testing"
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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      newDataResource("not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				PreConfig: func() {
					db := openDBConnection()
					defer db.Close()

					_, err := db.Exec(fmt.Sprintf("CREATE DATABASE [%s] COLLATE %s", dbSettings.Name, dbSettings.Collation))
					if err != nil {
						t.Fatal(err)
					}

					err = db.QueryRow("SELECT DB_ID(@p1)", dbSettings.Name).Scan(&dbId)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: newDataResource("data_test_db"),
				Check: resource.ComposeAggregateTestCheckFunc(
					//resource.TestCheckResourceAttrWith(resourceName, "id", func(value string) error {
					//	assert.Equal(t, dbId, value, "db_id")
					//	return nil
					//}),
					resource.TestCheckResourceAttrPtr(resourceName, "id", &dbId),
					resource.TestCheckResourceAttr(resourceName, "name", dbSettings.Name),
					resource.TestCheckResourceAttr(resourceName, "collation", dbSettings.Collation),
				),
			},
		},
	})
}
