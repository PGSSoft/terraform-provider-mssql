data "mssql_database" "example" {
  name = "example"
}

data "mssql_sql_user" "example" {
  name        = "example_user"
  database_id = data.mssql_database.example.id
}

data "mssql_schema" "example" {
  name        = "example_schema"
  database_id = data.mssql_database.example.id
}

data "mssql_schema_permissions" "example" {
  schema_id    = data.mssql_schema.example.id
  principal_id = data.mssql_sql_user.example.id
}

output "permissions" {
  value = data.mssql_schema_permissions.example.permissions
}