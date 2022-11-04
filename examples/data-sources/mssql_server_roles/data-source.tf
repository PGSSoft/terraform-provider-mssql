data "mssql_server_roles" "all" {}

output "roles" {
  value = data.mssql_server_roles.all.roles
}