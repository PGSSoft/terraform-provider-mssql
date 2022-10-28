data "mssql_database" "example" {
  name = "example"
}

data "mssql_sql_user" "example" {
  name        = "example_user"
  database_id = data.mssql_database.example.id
}

data "mssql_database_permissions" "example" {
  principal_id = data.mssql_sql_user.example.id
}

output "permissions" {
  value = data.mssql_database_permissions.example.permissions
}