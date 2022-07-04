package provider

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestDatabaseRoleData(t *testing.T) {
	createDB(t, "db_role_data_test")

	newDataResource := func(resourceName string, roleName string) string {
		return fmt.Sprintf(`
data "mssql_database" %[1]q {
	name = "db_role_data_test"
}

data "mssql_database_role" %[1]q {
	name = %[2]q
	database_id = data.mssql_database.%[1]s.id
}
`, resourceName, roleName)
	}

	var roleResourceId, ownerResourceId string
	var dbId string

	setIds := func(roleId int, ownerId int) {
		formatId := func(id int) string { return fmt.Sprintf("%s/%d", dbId, id) }
		roleResourceId = formatId(roleId)
		ownerResourceId = formatId(ownerId)
	}

	attributesCheck := func(resourceName string) resource.TestCheckFunc {
		resourceName = fmt.Sprintf("data.mssql_database_role.%s", resourceName)
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttrPtr(resourceName, "id", &roleResourceId),
			resource.TestCheckResourceAttrPtr(resourceName, "owner_id", &ownerResourceId),
			resource.TestCheckResourceAttrPtr(resourceName, "database_id", &dbId),
		)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      newDataResource("not_exists", "not_exists"),
				ExpectError: regexp.MustCompile("not exist"),
			},
			{
				PreConfig: func() {
					withDBConnection(func(conn *sql.DB) {
						var roleId, ownerId int

						err := conn.QueryRow(`
USE [db_role_data_test];
CREATE ROLE [test_owner];
CREATE ROLE [test_role] AUTHORIZATION [test_owner];
SELECT DATABASE_PRINCIPAL_ID('test_role'), DATABASE_PRINCIPAL_ID('test_owner'), DB_ID();
`).Scan(&roleId, &ownerId, &dbId)

						require.NoError(t, err, "creating role")
						setIds(roleId, ownerId)
					})
				},
				Config: newDataResource("exists", "test_role"),
				Check:  attributesCheck("exists"),
			},
			{
				Config: `
data "mssql_database_role" "master" {
	name = "public"
}
`,
				Check: resource.ComposeTestCheckFunc(
					sqlCheck(func(db *sql.DB) error {
						var roleId, ownerId int
						err := db.QueryRow("SELECT DB_ID(), DATABASE_PRINCIPAL_ID('public'), DATABASE_PRINCIPAL_ID('dbo')").Scan(&dbId, &roleId, &ownerId)
						setIds(roleId, ownerId)
						return err
					}),
					attributesCheck("master"),
				),
			},
		},
	})
}
