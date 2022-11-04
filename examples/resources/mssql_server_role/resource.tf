resource "mssql_server_role" "owner" {
  name = "owner_role"
}

resource "mssql_server_role" "example" {
  name     = "example"
  owner_id = mssql_server_role.owner.id
}