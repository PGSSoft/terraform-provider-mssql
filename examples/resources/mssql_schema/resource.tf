data "mssql_database" "example" {
  name = "example"
}

data "mssql_sql_user" "owner" {
  name = "example_user"
}

resource "mssql_schema" "example" {
  name        = "example"
  database_id = data.mssql_database.example.id
  owner_id    = data.mssql_sql_user.owner.id
}