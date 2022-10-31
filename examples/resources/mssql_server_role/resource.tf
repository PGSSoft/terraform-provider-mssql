resource "mssql_server_role" "owner" {
  name = "owner_role"
}

resource "mssql_database_role" "example" {
  name     = "example"
  owner_id = data.mssql_server_role.owner.id
}