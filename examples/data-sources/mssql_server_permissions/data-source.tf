data "mssql_sql_login" "example" {
  name = "example_login"
}

data "mssql_server_permissions" "example" {
  principal_id = data.mssql_sql_login.example.principal_id
}

output "permissions" {
  value = data.mssql_server_permissions.example.permissions
}