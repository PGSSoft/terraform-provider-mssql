data "mssql_database" "master" {
  name = "master"
}

data "mssql_database_roles" "example" {
  database_id = data.mssql_database.master.id
}

output "roles" {
  value = data.mssql_database_roles.example.roles
}
