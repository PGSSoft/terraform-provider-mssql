package provider

import (
	"database/sql"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDatabaseRoleListData(t *testing.T) {
	createDB(t, "db_role_list_test")

	var roleResourceId, ownerResourceId string
	var dbId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: newProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "mssql_database_roles" "master" {}`,
				Check: resource.TestCheckTypeSetElemNestedAttrs("data.mssql_database_roles.master", "roles.*", map[string]string{
					"id":          "1/0",
					"name":        "public",
					"database_id": "1",
					"owner_id":    "1/1",
				}),
			},
			{
				PreConfig: func() {
					withDBConnection(func(conn *sql.DB) {
						var roleId, ownerId int
						err := conn.QueryRow(`
USE [db_role_list_test];
CREATE ROLE test_owner;
CREATE ROLE test_role AUTHORIZATION test_owner;
SELECT DB_ID(), DATABASE_PRINCIPAL_ID('test_role'), DATABASE_PRINCIPAL_ID('test_owner');
`).Scan(&dbId, &roleId, &ownerId)

						require.NoError(t, err, "creating role")

						roleResourceId = fmt.Sprintf("%s/%d", dbId, roleId)
						ownerResourceId = fmt.Sprintf("%s/%d", dbId, ownerId)
					})
				},
				Config: `
data "mssql_database" "test" {
	name = "db_role_list_test"
}

data "mssql_database_roles" "test" {
	database_id = data.mssql_database.test.id
}
`,
				Check: func(state *terraform.State) error {
					return resource.TestCheckTypeSetElemNestedAttrs("data.mssql_database_roles.test", "roles.*", map[string]string{
						"id":          roleResourceId,
						"name":        "test_role",
						"database_id": dbId,
						"owner_id":    ownerResourceId,
					})(state)
				},
			},
		},
	})
}
