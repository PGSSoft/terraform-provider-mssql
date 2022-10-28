data "mssql_database" "example" {
  name = "example"
}

data "mssql_sql_user" "example" {
  name        = "example_user"
  database_id = data.mssql_database.example.id
}

resource "mssql_database_permission" "delete_to_example" {
  principal_id = data.mssql_sql_user.example.id
  permission   = "DELETE"
}